package bigquery

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/cenkalti/backoff"
	bigquery "google.golang.org/api/bigquery/v2"
)

type tableStreamer struct {
	streamer *Streamer
	service  *bigquery.TabledataService
	table    string
	suffix   func() string

	incoming chan interface{}
	stop     chan struct{}

	queue  []row
	lastID int64

	flushInterval time.Duration
	flushMax      int
	crankiness    *backoff.ExponentialBackOff
}

func newTableStreamer(streamer *Streamer, table string, suffix func() string) *tableStreamer {
	ts := &tableStreamer{
		streamer: streamer,
		service:  bigquery.NewTabledataService(streamer.service),
		table:    table,
		suffix:   suffix,

		incoming: make(chan interface{}, bufferSize),
		stop:     make(chan struct{}),

		lastID: rand.Int63(),

		flushInterval: 10 * time.Second,
		flushMax:      bufferSize,
		crankiness:    backoff.NewExponentialBackOff(),
	}
	ts.crankiness.MaxElapsedTime = 0
	ts.crankiness.InitialInterval = 2 * time.Second
	ts.crankiness.NextBackOff()
	return ts
}

func (ts *tableStreamer) insert(data interface{}) {
	ts.incoming <- data
}

type row struct {
	id    string
	val   map[string]bigquery.JsonValue
	iface interface{}
	tries int
}

func (ts *tableStreamer) newRow(v interface{}) (row, error) {
	encoded, err := Encode(v)
	if err != nil {
		return row{}, err
	}
	return row{
		id:    hostname + strconv.FormatInt(ts.nextID(), 36),
		val:   encoded,
		iface: v,
	}, nil
}

func (ts *tableStreamer) nextID() int64 {
	ts.lastID++
	return ts.lastID
}

func (ts *tableStreamer) run() {
	tick := time.NewTicker(ts.flushInterval)
	defer tick.Stop()
	for {
		select {
		case data := <-ts.incoming:
			r, err := ts.newRow(data)
			if err != nil {
				ts.streamer.Errors <- err
				continue
			}
			ts.queue = append(ts.queue, r)
			if len(ts.queue) >= ts.flushMax {
				ts.flush()
			}
		case <-tick.C:
			ts.flush()
		case <-ts.stop:
			// should be flushed by Stop
			return
		}
	}
}

func (ts *tableStreamer) flush() {
	if len(ts.queue) == 0 {
		return
	}

	rows := make([]*bigquery.TableDataInsertAllRequestRows, 0, len(ts.queue))
	for _, row := range ts.queue {
		rows = append(rows, &bigquery.TableDataInsertAllRequestRows{
			InsertId: row.id,
			Json:     row.val,
		})
	}

	//  send request
	request := &bigquery.TableDataInsertAllRequest{
		Kind: "bigquery#tableDataInsertAllRequest",
		Rows: rows,
	}
	if ts.suffix != nil {
		request.TemplateSuffix = ts.suffix()
	}

	resp, err := ts.service.InsertAll(ts.streamer.project, ts.streamer.dataset, ts.table, request).Do()

	// success
	if err == nil {
		var nextQueue []row
		if len(resp.InsertErrors) > 0 {
			for _, errs := range resp.InsertErrors {
				for _, err := range errs.Errors {
					r := ts.queue[errs.Index]
					r.tries++
					if r.tries < maxInsertTries {
						nextQueue = append(nextQueue, r)
					} else {
						ts.streamer.Errors <- fmt.Errorf("BQ insert error: %v", err.Reason)
					}
				}
			}
		}
		// TODO: figure out how to deal w/ row errors
		// TODO: schema changes...
		if len(nextQueue) > 0 {
			log.Printf("[%s] Sent: %d, Next: %d", ts.table, len(ts.queue), len(nextQueue))
		}
		ts.queue = nextQueue
		return
	}

	// internal errors
	if gerr, ok := err.(*googleapi.Error); ok {
		switch gerr.Code {
		case 500, 503:
			log.Println("BQ: Internal error:", gerr)
			// sleep & retry
			time.Sleep(ts.crankiness.NextBackOff())
			return
		}
	}

	// missing table
	if ts.streamer.CreateTables && isTableNotFoundErr(err) {
		schema, _ := Schema(ts.queue[0].iface)
		if makeTableErr := ts.createTable(schema); makeTableErr == nil {
			wait := ts.crankiness.NextBackOff()
			log.Printf("Made table %s, retrying after %v...", ts.table, wait)
			time.Sleep(wait)
			return
		} else {
			ts.streamer.Errors <- makeTableErr
			// try again
			time.Sleep(ts.crankiness.NextBackOff())
			return
		}
	}

	// some other kind of unexpected error
	// keep trying
	if err != nil {
		ts.streamer.Errors <- err
	} else {
		ts.crankiness.Reset()
	}
}

func (ts *tableStreamer) createTable(schema *bigquery.TableSchema) error {
	tablesService := bigquery.NewTablesService(ts.streamer.service)
	table := &bigquery.Table{
		Schema: schema,
		TableReference: &bigquery.TableReference{
			ProjectId: ts.streamer.project,
			DatasetId: ts.streamer.dataset,
			TableId:   ts.table,
		},
	}
	_, err := tablesService.Insert(ts.streamer.project, ts.streamer.dataset, table).Do()
	if err == nil || isAlreadyExistsErr(err) {
		return nil
	}
	return err
}

package bigquery

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/googleapi"

	bigquery "google.golang.org/api/bigquery/v2"
)

func NewStreamer(service *bigquery.Service, project, dataset string) *Streamer {
	return &Streamer{
		service: service,
		project: project,
		dataset: dataset,

		tables: make(map[string]*tableStreamer),
		Errors: make(chan error, bufferSize),
	}
}

func (s *Streamer) Insert(table string, data interface{}, suffix func() string) {
	s.mu.RLock()
	ts := s.tables[table]
	s.mu.RUnlock()

	if ts == nil {
		s.mu.Lock()
		ts = newTableStreamer(s, table, suffix)
		s.tables[table] = ts
		go ts.run()
		s.mu.Unlock()
	}

	ts.insert(data)
}

func (s *Streamer) Stop() {
	s.mu.Lock()
	for table, ts := range s.tables {
		close(ts.stop)
		delete(s.tables, table)
		ts.flush()
	}
	s.mu.Unlock()
}

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

	"golang.org/x/oauth2"
	// "golang.org/x/oauth2/google"
	"github.com/cenkalti/backoff"
	"golang.org/x/oauth2/jwt"
	bigquery "google.golang.org/api/bigquery/v2"

	"google.golang.org/api/googleapi"
)

const (
	bufferSize     = 15000
	maxInsertTries = 10
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

type BigQuery struct {
	project      string
	dataset      string
	service      *bigquery.Service
	mu           sync.RWMutex
	createTables bool
	errors       chan error
}

func NewBigQueryService(c *jwt.Config) (service *bigquery.Service, err error) {
	client := c.Client(oauth2.NoContext)
	service, err = bigquery.New(client)
	return
}

func isTableNotFoundErr(err error) bool {
	if gerr, ok := err.(*googleapi.Error); ok {
		if gerr.Code == 404 && strings.Contains(gerr.Message, "Not found: Table") {
			return true
		}
	}
	return false
}

func isAlreadyExistsErr(err error) bool {
	if gerr, ok := err.(*googleapi.Error); ok {
		if gerr.Code == 409 && strings.Contains(gerr.Message, "Already Exists") {
			return true
		}
	}
	return false
}

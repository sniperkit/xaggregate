package github

import (
	"net/http"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/segmentio/stats/httpstats"
)

var (
	CacheEngine     = "badger"
	CachePrefixPath = "./shared/data/cache/http"
	xcache          httpcache.Cache
	xtransport      *httpcache.Transport
)

func initCacheTransport() (httpcache.Cache, *httpcache.Transport) {
	defer funcTrack(time.Now())

	backendCache, err := newCacheBackend(CacheEngine, CachePrefixPath)
	if err != nil {
		log.Fatal("cache err", err.Error())
	}

	var httpTransport = http.DefaultTransport
	httpTransport = httpstats.NewTransport(httpTransport)
	http.DefaultTransport = httpTransport

	cachingTransport := httpcache.NewTransportFrom(backendCache, httpTransport) // httpcache.NewMemoryCacheTransport()
	cachingTransport.MarkCachedResponses = true

	return backendCache, cachingTransport
}

func setCacheExpire(key string, date time.Time) bool {
	defer funcTrack(time.Now())

	return true
}

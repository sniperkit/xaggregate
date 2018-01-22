package github

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/segmentio/stats"
	"github.com/segmentio/stats/httpstats"

	// "github.com/lalyos/httptrace"

	"github.com/gregjones/httpcache"
	"github.com/sniperkit/xcache/backend/default/badger"
	"github.com/sniperkit/xcache/backend/default/diskv"

	"github.com/sniperkit/xapi/pkg"
	"github.com/sniperkit/xapi/service/github"

	"github.com/sniperkit/xtask/util/fs" // move into a separate repo/package
)

var (
	statsEngine               *stats.Engine
	revalidationDefaultMaxAge = &githubproxy.MaxAge{
		User:         time.Hour * 24 * 7,
		Repository:   time.Hour * 24 * 7,
		Ref:          time.Hour * 24 * 7,
		Tree:         time.Hour * 24 * 7,
		Readme:       time.Hour * 24 * 7,
		Lang:         time.Hour * 24 * 7,
		Topic:        time.Hour * 24 * 7,
		Repositories: time.Hour * 24 * 7,
		Activity:     time.Hour * 12 * 7,
		Starred:      time.Hour * 12 * 7,
	}
)

func newCacheBackend(engine string, prefixPath string) (backend httpcache.Cache, err error) {
	defer funcTrack(time.Now())

	fsutil.EnsureDir(prefixPath)
	engine = strings.ToLower(engine)

	switch CacheEngine {
	case "diskv":
		cacheStoragePrefixPath := filepath.Join(prefixPath, "cacher.diskv")
		fsutil.EnsureDir(cacheStoragePrefixPath)
		backend = diskcache.New(cacheStoragePrefixPath)

	case "badger":
		cacheStoragePrefixPath := filepath.Join(prefixPath, "cacher.badger")
		fsutil.EnsureDir(cacheStoragePrefixPath)
		backend, err = badgercache.New(
			&badgercache.Config{
				ValueDir:    "api.github.com.v3.snappy",
				StoragePath: cacheStoragePrefixPath,
				SyncWrites:  false,
				Debug:       false,
				Compress:    true,
			})

	case "memory":
		backend = httpcache.NewMemoryCache()

	default:
		backend = nil
	}

	return
}

func newCacheTransport(c httpcache.Cache, markCachedResponses bool) http.RoundTripper {
	defer funcTrack(time.Now())

	t := httpcache.NewTransport(c)
	t.MarkCachedResponses = markCachedResponses
	return t
}

func newCacheRevalidationTransport(rt http.RoundTripper, config *githubproxy.MaxAge) *apiproxy.RevalidationTransport {
	defer funcTrack(time.Now())

	return &apiproxy.RevalidationTransport{
		Transport: rt,
		Check: (&githubproxy.MaxAge{
			User:         time.Hour * 24,
			Repository:   time.Hour * 24,
			Repositories: time.Hour * 24,
			Activity:     time.Hour * 12,
		}).Validator(),
	}
}

func newStatsTransport(rt http.RoundTripper) http.RoundTripper {
	defer funcTrack(time.Now())

	return httpstats.NewTransport(rt)
}

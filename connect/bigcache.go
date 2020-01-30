package connect

import (
	"github.com/allegro/bigcache"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	bigCache    *bigcache.BigCache
	once        sync.Once
	cacheConfig = bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 256,
		// time after which entry can be evicted
		LifeWindow: 1 * time.Minute,

		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed.
		// Setting to < 1 second is counterproductive — bigcache has a one second resolution.
		CleanWindow: 2 * time.Minute,

		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 256,

		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 256000,

		// prints information about additional memory allocation
		Verbose: true,

		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: 256,

		// callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A bitmask representing the reason will be returned.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		OnRemove: nil,

		// OnRemoveWithReason is a callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A constant representing the reason will be passed through.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		// Ignored if OnRemove is specified.
		OnRemoveWithReason: nil,
	}
)

func ConnectBigcache() (*bigcache.BigCache, error) {
	once.Do(func() {
		var err error
		bigCache, err = bigcache.NewBigCache(cacheConfig)
		if err != nil {
			log.Error("connect big cache fail: %s", err)
		}
	})

	return bigCache, nil
}

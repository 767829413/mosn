package store

import (
	"sync"
	"sync/atomic"
	"time"

	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/types"
	"mosn.io/mosn/pkg/utils"
)

func writeCacheConfig() {
	if getCache() {
		b, err := Dump()
		if err != nil {
			log.DefaultLogger.Alertf(types.ErrorKeyConfigDump, "dump cache config failed, caused by: "+err.Error())
		}
		if err := utils.WriteFileSafety(types.MosnCacheConfig, b, 0644); err != nil {
			log.DefaultLogger.Alertf(types.ErrorKeyConfigDump, "dump cache config failed, caused by: "+err.Error())
		}
	}
}

var caching int32

func setCache() {
	atomic.CompareAndSwapInt32(&caching, 0, 1)
}

func getCache() bool {
	return atomic.CompareAndSwapInt32(&caching, 1, 0)
}

var once sync.Once

// TODO: if we consider smooth upgrade, add a lock here and call lock in reconfigure
func CacheConfigHandler() {
	once.Do(func() {
		for {
			time.Sleep(3 * time.Second)
			writeCacheConfig()
		}
	})
}

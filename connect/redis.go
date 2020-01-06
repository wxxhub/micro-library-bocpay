package connect

import (
	"micro-library/helper"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var rds *Rds

type Rds struct {
	sync.RWMutex
	Map map[string]*redis.ClusterClient
}

func init() {
	rds = new(Rds)
	rds.Map = make(map[string]*redis.ClusterClient)
}

func ConnectRedis(ctx context.Context, hlp *helper.Helper, srvName string, name string) (*redis.ClusterClient, error) {
	timer := hlp.Timer
	timer.Start("connectRedis")
	defer timer.End("connectRedis")

	rds.RLock()
	rd, ok := rds.Map[name]
	rds.RUnlock()
	if !ok {
		conf, watcher, err := ConnectConfig(srvName, "redis")
		if err != nil {
			RedisLog.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("read redis config fail")
			return nil, fmt.Errorf("read redis config fail: %w", err)
		}

		var clusterConfig redis.ClusterOptions
		conf.Get(srvName, "redis", name).Scan(&clusterConfig)

		rd = redis.NewClusterClient(&clusterConfig)

		pong, err := rd.Ping().Result()
		if err != nil {
			RedisLog.WithFields(logrus.Fields{
				"addr":  clusterConfig.Addrs,
				"pong":  pong,
				"error": err.Error(),
			}).Error("connect redis fail")
			return nil, fmt.Errorf("connect redis fail: %w", err)
		}
		rds.Lock()
		rds.Map[name] = rd
		rds.Unlock()

		go func() {
			v, err := watcher.Next()
			if err != nil {
				RedisLog.WithFields(logrus.Fields{
					"error": err,
					"name":  name,
					"file":  string(v.Bytes()),
				}).Warn("reconect redis")
			} else {
				RedisLog.WithFields(logrus.Fields{
					"name": name,
					"file": string(v.Bytes()),
				}).Info("reconnect redis")

				//配置更新了，释放所有已有的rd对象，关闭连接
				rds.RLock()
				rd, ok := rds.Map[name]
				rds.RUnlock()
				if ok {
					rds.Lock()
					delete(rds.Map, name)
					rds.Unlock()
				}
				//10秒后，关闭旧的redis连接
				time.Sleep(time.Duration(10) * time.Second)
				err = rd.Close()
				if err == nil {
					RedisLog.WithFields(logrus.Fields{
						"name": name,
						"file": string(v.Bytes()),
					}).Info("close rds")
				} else {
					RedisLog.WithFields(logrus.Fields{
						"error": err,
						"name":  name,
						"file":  string(v.Bytes()),
					}).Warn("close rds error")
				}
			}
			return
		}()
	}
	newRedis := rd.WithContext(ctx)
	return newRedis, nil
}

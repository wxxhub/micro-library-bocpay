package connect

import (
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"time"
)

type RedisStats struct {
	Hits     prometheus.Gauge // number of times free connection was found in the pool
	Misses   prometheus.Gauge // number of times free connection was NOT found in the pool
	Timeouts prometheus.Gauge // number of times a wait timeout occurred

	TotalConns prometheus.Gauge // number of total connections in the pool
	IdleConns  prometheus.Gauge // number of idle connections in the pool
	StaleConns prometheus.Gauge // number of stale connections removed from the pool
}

func newRedisStats(srvName string, isCluster bool) *RedisStats {
	var prefix string
	if isCluster {
		prefix = "cluster_"
	} else {
		prefix = "client_"
	}
	prefix += srvName + "_redis_"

	result := &RedisStats{
		Hits: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "hits",
		}),
		Misses: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "misses",
		}),
		Timeouts: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "timeouts",
		}),
		TotalConns: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "totalconnnes",
		}),
		IdleConns: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "idleconns",
		}),
		StaleConns: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "staleconnes",
		}),
	}

	x := reflect.ValueOf(*result)

	for i := 0; i < x.NumField(); i++ {
		_ = prometheus.Register(x.Field(i).Interface().(prometheus.Gauge))
	}

	return result
}

type RdsCollector struct {
	Cluster map[string]*RedisStats
	Client  map[string]*RedisStats
}

var rsc *RdsCollector

func init() {
	rsc = new(RdsCollector)
	rsc.Cluster = make(map[string]*RedisStats)
	rsc.Client = make(map[string]*RedisStats)
	go func() {
		for range time.Tick(15 * time.Second) {
			redisMetrics()
		}
	}()
}

func redisMetrics() {
	if rds == nil {
		return
	}
	rds.RLock()
	rds.RUnlock()

	for srvName, cluster := range rds.Map {
		stats := cluster.PoolStats()
		if _, ok := rsc.Cluster[srvName]; !ok {
			rsc.Cluster[srvName] = newRedisStats(srvName, true)
		}

		rsc.Cluster[srvName].Hits.Set(float64(stats.Hits))
		rsc.Cluster[srvName].IdleConns.Set(float64(stats.IdleConns))
		rsc.Cluster[srvName].Misses.Set(float64(stats.Misses))
		rsc.Cluster[srvName].StaleConns.Set(float64(stats.StaleConns))
		rsc.Cluster[srvName].Timeouts.Set(float64(stats.Timeouts))
		rsc.Cluster[srvName].TotalConns.Set(float64(stats.TotalConns))
	}

	for srvName, client := range rds.MapRedis {
		stats := client.PoolStats()
		if _, ok := rsc.Client[srvName]; !ok {
			rsc.Client[srvName] = newRedisStats(srvName, false)
		}

		rsc.Client[srvName].Hits.Set(float64(stats.Hits))
		rsc.Client[srvName].IdleConns.Set(float64(stats.IdleConns))
		rsc.Client[srvName].Misses.Set(float64(stats.Misses))
		rsc.Client[srvName].StaleConns.Set(float64(stats.StaleConns))
		rsc.Client[srvName].Timeouts.Set(float64(stats.Timeouts))
		rsc.Client[srvName].TotalConns.Set(float64(stats.TotalConns))
	}
}

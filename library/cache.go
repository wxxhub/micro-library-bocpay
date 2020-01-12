package library

import (
	"github.com/lifenglin/micro-library/connect"
	"github.com/lifenglin/micro-library/helper"
	"context"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"time"
)

func GetCache(ctx context.Context, hlp *helper.Helper, srvName string, name string, redisKey string, value interface{}) (err error) {
	log := hlp.RedisLog
	redis, err := connect.ConnectRedis(ctx, hlp, srvName, name)
	if err != nil {
		return err
	}
	redisTTL, err := redis.TTL(redisKey).Result()
	var bytes []byte
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":    err,
			"redisKey": redisKey,
		}).Warn("getTTLFromRedis error")
		return err
	} else {
		if -1 == redisTTL.Seconds() {
			log.WithFields(logrus.Fields{
				"error":    err,
				"redisKey": redisKey,
			}).Warn("ttl should not no expire")
			return err
		} else if -2 == redisTTL.Seconds() {
			//empty
		} else {
			bytes, err = redis.Get(redisKey).Bytes()
			if err != nil && err.Error() != "redis: nil" {
				log.WithFields(logrus.Fields{
					"error":    err,
					"redisKey": redisKey,
				}).Warn("getDataFromRedis error")
			} else if err != nil {
				//缓存未命中，从数据库中获取数据
				log.WithFields(logrus.Fields{
					"redisKey": redisKey,
					"bytes":    string(bytes),
				}).Trace("miss cache")
				return err
			}
		}
	}
	//如果命中缓存，则从缓存中拿出数据返回
	if bytes != nil {
		err := json.Unmarshal(bytes, value)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error":    err,
				"redisKey": redisKey,
			}).Warn("json unmarshal error")
			return err
		}
		log.WithFields(logrus.Fields{
			"redisKey": redisKey,
			"value":    value,
			"bytes":    string(bytes),
		}).Trace("all hit cache")
		return nil
	}
	return errors.New("redis: nil")
}


func MgetCache(ctx context.Context, hlp *helper.Helper, srvName string, name string, redisKey []string, value []interface{}) (err error) {
	if len(redisKey) != len(value) {
		return errors.New("len is not eq")
	}
	for key, item := range redisKey {
		GetCache(ctx, hlp, srvName, name, item, value[key])
	}
	return nil
}

func MsetCache(ctx context.Context, hlp *helper.Helper, srvName string, name string, redisKey []string, value []interface{}, expire time.Duration) (err error) {
	if len(redisKey) != len(value) {
		return errors.New("len is not eq")
	}
	for key, item := range redisKey {
		SetCache(ctx, hlp, srvName, name, item, value[key], expire)
	}
	return nil
}

func GetCacheNum(ctx context.Context, hlp *helper.Helper, srvName string, name string, redisKey string) (num int64, err error) {
	log := hlp.RedisLog
	redis, err := connect.ConnectRedis(ctx, hlp, srvName, name)
	if err != nil {
		return num, err
	}
	redisTTL, err := redis.TTL(redisKey).Result()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":    err,
			"redisKey": redisKey,
		}).Warn("getTTLFromRedis error")
		return num, errors.New("redis: nil")
	} else {
		if -1 == redisTTL.Seconds() {
			log.WithFields(logrus.Fields{
				"error":    err,
				"redisKey": redisKey,
			}).Warn("ttl should not no expire")
			return num, errors.New("redis: nil")
		} else if -2 == redisTTL.Seconds() {
			//empty
			return num, errors.New("redis: nil")
		} else {
			num, err = redis.Get(redisKey).Int64()
			if err != nil && err.Error() != "redis: nil" {
				log.WithFields(logrus.Fields{
					"error":    err,
					"redisKey": redisKey,
				}).Warn("getDataFromRedis error")
				return num, errors.New("redis: nil")
			} else if err != nil {
				//缓存未命中，从数据库中获取数据
				log.WithFields(logrus.Fields{
					"redisKey": redisKey,
					"err":      err,
					"num":      num,
				}).Trace("miss cache")
				return num, errors.New("redis: nil")
			}
		}
	}
	//如果命中缓存，则从缓存中拿出数据返回
	log.WithFields(logrus.Fields{
		"redisKey": redisKey,
		"num":      num,
	}).Trace("all hit cache")
	return num, nil
}

func DelCache(ctx context.Context, hlp *helper.Helper, srvName string, name string, redisKey string) (err error) {
	log := hlp.RedisLog
	redis, err := connect.ConnectRedis(ctx, hlp, srvName, name)
	if err != nil {
		return err
	}
	err = redis.Del(redisKey).Err()
	log.WithFields(logrus.Fields{
		"error":    err,
		"redisKey": redisKey,
	}).Trace("del redis")
	if nil != err {
		log.WithFields(logrus.Fields{
			"error":    err,
			"redisKey": redisKey,
		}).Warn("del error")
		return err
	}
	return nil
}

func SetCache(ctx context.Context, hlp *helper.Helper, srvName string, name string, redisKey string, value interface{}, expire time.Duration) (err error) {
	log := hlp.RedisLog
	redis, err := connect.ConnectRedis(ctx, hlp, srvName, name)
	if err != nil {
		return err
	}

	redisBytes, err := json.Marshal(value)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":    err,
			"redisKey": redisKey,
		}).Warn("json marshal error")
	}
	err = redis.Set(redisKey, redisBytes, expire).Err()
	if err != nil {
		log.WithFields(logrus.Fields{
			"redisKey": redisKey,
			"error":    err,
			"value":    value,
			"expire":   expire,
		}).Warn("setRedis error")
		return err
	} else {
		log.WithFields(logrus.Fields{
			"redisKey": redisKey,
			"value":    value,
			"string":   string(redisBytes),
			"expire":   expire,
		}).Trace("set redis")
	}
	return nil
}

func SetCacheNum(ctx context.Context, hlp *helper.Helper, srvName string, name string, redisKey string, value int64, expire time.Duration) (err error) {
	log := hlp.RedisLog
	redis, err := connect.ConnectRedis(ctx, hlp, srvName, name)
	if err != nil {
		return err
	}

	err = redis.Set(redisKey, value, expire).Err()
	if err != nil {
		log.WithFields(logrus.Fields{
			"redisKey": redisKey,
			"error":    err,
			"value":    value,
			"expire":   expire,
		}).Warn("setRedis error")
		return err
	} else {
		log.WithFields(logrus.Fields{
			"redisKey": redisKey,
			"value":    value,
			"expire":   expire,
		}).Trace("set redis")
	}
	return nil
}

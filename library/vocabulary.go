package library

import (
	"github.com/lifenglin/micro-library/connect"
	"github.com/lifenglin/micro-library/helper"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"time"
)

type Vocabulary struct {
	Id         uint   `json:"id"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
	Params     string `json:"params"`
	Status     uint   `json:"status"`
	Key        string `json:"key"`
	UpdateTime uint   `json:"update_time"`
	OpOrAd     string `json:"op_or_ad"`
}

func GetVocabularyListByKey(ctx context.Context, hlp *helper.Helper, key string) (map[uint]Vocabulary, error) {
	conf, _, err := connect.ConnectConfig("Vocabulary", filepath.Join("data", key))
	Log := hlp.Log
	if err != nil {
		Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("read vocabulary config fail")
		return nil, fmt.Errorf("read vocabulary config fail: %w", err)
	}
	var list map[uint]Vocabulary
	conf.Get("Vocabulary", "data", key).Scan(&list)
	for key, item := range list {
		if item.Status == 1 {
			//下线词表，干掉
			delete(list, key)
			continue
		}
		//设置了生效时间
		if false == (item.Start == 0 && item.End == 0) {
			if false == (item.Start <= time.Now().Unix() && time.Now().Unix() <= item.End) {
				//过期词表，干掉
				delete(list, key)
				continue
			}
		}
	}
	return list, nil
}

func GetVocabularyParamsByKey(ctx context.Context, hlp *helper.Helper, key string) (map[uint]map[string]interface{}, error) {
	Log := hlp.Log
	list, err := GetVocabularyListByKey(ctx, hlp, key)
	if err != nil {
		Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("read vocabulary config fail")
		return nil, fmt.Errorf("read vocabulary config fail: %w", err)
	}
	paramsList := make(map[uint]map[string]interface{})
	for _, item := range list {
		it := new(map[string]interface{})
		json.Unmarshal([]byte(item.Params), it)
		paramsList[item.Id] = *it
	}
	return paramsList, nil
}

func GetVocabularyParamsByKeyAndId(ctx context.Context, hlp *helper.Helper, key string, id uint) (map[string]interface{}, error) {
	Log := hlp.Log
	paramsList, err := GetVocabularyParamsByKey(ctx, hlp, key)

	if err != nil {
		Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("read vocabulary config fail")
		return nil, fmt.Errorf("read vocabulary config fail: %w", err)
	}
	_, ok := paramsList[id]
	if ok {
		return paramsList[id], nil
	}
	return nil, fmt.Errorf("vocabulary config is nil: %w", err)
}

func GetWhiteListUid(ctx context.Context, hlp *helper.Helper, log *logrus.Entry, id uint) (uids []uint32) {
	officialResult, err := GetVocabularyParamsByKeyAndId(ctx, hlp,
		"nice_white_list", id)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Warn("get vocabulary white list fail")
	} else {
		if _, ok := officialResult["uid"]; !ok {
			log.WithFields(logrus.Fields{
				"error": errors.New("not found uid"),
			}).Warn("get vocabulary uid key fail")
		} else {
			for _, uid := range officialResult["uid"].([]interface{}) {
				uint64Uid, err := strconv.ParseUint(uid.(string), 10, 32)
				if err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
					}).Warn("parse uint failed")
					continue
				}
				uids = append(uids, uint32(uint64Uid))
			}
		}
	}
	return
}

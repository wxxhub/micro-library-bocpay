package connect

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/lifenglin/micro-library/helper"
	"github.com/sirupsen/logrus"
	"sync"
)

var sqlites *Sqlites

type Sqlites struct {
	sync.RWMutex
	Map map[string]*gorm.DB
}

func init() {
	sqlites = new(Sqlites)
	sqlites.Map = make(map[string]*gorm.DB)
}

func ConnectSqlite(ctx context.Context, hlp *helper.Helper, srvName string, name string) (*gorm.DB, error) {
	timer := hlp.Timer
	timer.Start("connectSqlite")
	defer timer.End("connectSqlite")

	sqlitesKey := name
	mysqlLog := hlp.MysqlLog
	var err error
	sqlites.RLock()
	db, ok := sqlites.Map[sqlitesKey]
	sqlites.RUnlock()
	if !ok {
		sqlites.Lock()
		existDb, ok := sqlites.Map[sqlitesKey]
		if ok {
			db = existDb
		} else {
			db, err = gorm.Open("sqlite3", ":memory:")
			if err != nil {
				mysqlLog.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("connect sqlite fail")
				sqlites.Unlock()
				return nil, fmt.Errorf("connect sqlite fail: %w", err)
			}
			db.SingularTable(true)
			db.BlockGlobalUpdate(false)
			sqlites.Map[sqlitesKey] = db
		}
		sqlites.Unlock()
	}
	newDb := db.New()
	newDb.SetLogger(mysqlLog)
	conf, _, err := ConnectConfig(srvName, "log")
	if err != nil {
		//配置获取失败
		mysqlLog.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("read log config fail")
	} else {
		newDb.LogMode(conf.Get(srvName, "log", "mysql_detailed_log").Bool(false))
	}
	return newDb, nil
}

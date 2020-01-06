package helper

import (
	"github.com/sirupsen/logrus"
	pool "gopkg.in/go-playground/pool.v3"
)

type Helper struct {
	Timer    *Timer
	Log      *logrus.Entry
	MysqlLog *logrus.Entry
	RedisLog *logrus.Entry
	Stat     map[string]int
	Pool     pool.Pool
}

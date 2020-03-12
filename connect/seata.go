package connect

import (
	"fmt"
	"github.com/lifenglin/micro-library/helper"
	"github.com/sirupsen/logrus"
	"math/rand"
	"seata.io/server/pkg/client"
	"time"
	"context"
)

type seataConfig struct {
	addrs []string
}

func ConnectSeata(ctx context.Context, hlp *helper.Helper, srvName string, name string, applicationID string, handler *client.ResourceHandler, resources []string) (*client.Client, error) {
	conf, _, err := ConnectConfig(srvName, "seata")
	if err != nil {
		hlp.RedisLog.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("read seata config fail")
		return nil, fmt.Errorf("read seata config fail: %w", err)
	}
	var seataConf seataConfig
	conf.Get(srvName, "seata").Scan(&seataConf)

	c := client.NewClient(client.Cfg{
		Addrs:             seataConf.addrs,
		HeartbeatDuration: time.Second * 5,
		Timeout:           time.Second * 10,
		ApplicationID:     applicationID,
		Version:           "0.5.0",
		Resources:         resources,
		Handler:           *handler,
		Seq:               rand.Uint64(),
	})
	return c, nil
}

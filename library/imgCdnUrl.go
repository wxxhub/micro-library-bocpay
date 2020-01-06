package library

import (
	"micro-library/connect"
	"micro-library/helper"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strings"
)

func GetImagCdnUrl(ctx context.Context, hlp *helper.Helper, relativePath string, suffix string) (string, error) {
	conf, _, err := connect.ConnectConfig("cdn", "hosts")
	Log := hlp.Log
	if err != nil {
		Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("read cdn config fail")
		return "", fmt.Errorf("read vocabulary config fail: %w", err)
	}
	type host struct {
		Host    string `json:"host"`
		Percent int    `json:"percent"`
	}
	var list []host
	conf.Get("cdn", "hosts", "image").Scan(&list)
	rint := rand.Intn(100)
	sum := 0
	usedHost := ""
	for _, item := range list {
		sum = sum + item.Percent
		if rint <= sum {
			usedHost = item.Host
			break
		}
	}
	return strings.TrimRight(usedHost, "/") + "/" + strings.TrimLeft(relativePath+suffix, "/"), nil
}

package library

import (
	"context"
	"github.com/lifenglin/micro-library/connect"
	"github.com/lifenglin/micro-library/helper"
)

func GetIncrementId(ctx context.Context, hlp *helper.Helper) (id uint64, err error) {
	redis, err := connect.ConnectIdGenerator(ctx, hlp)
	if err != nil {
		return 0, nil
	}
	//重试2次
	for i := 0; i < 2; i++ {
		id, err = redis.Do("getid").Uint64()
		if err == nil {
			return id, nil
		}
	}
	return 0, err

}

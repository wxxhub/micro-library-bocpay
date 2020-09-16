package connect

import (
	"errors"
	"github.com/lifenglin/micro-library/bocpay"
)

var bocpayClient *bocpay.Client

func ConnectBocpay(srvName string, confName string) (*bocpay.Client, error) {
	if bocpayClient != nil {
		return bocpayClient, nil
	}


	conf, _, err := ConnectConfig(srvName, confName)
	if err != nil {
		return nil, errors.New(srvName+" read bocpay config fail: "+err.Error())
	}

	config := new(bocpay.Config)

	// 直接解析会出现参数不全的情况，不采用
	//err = conf.Get(srvName, confName).Scan(config)

	config.Version = conf.Get(srvName, confName, "version").String("")
	if config.Version == "" {
		return nil, errors.New(srvName+", version is empty ")
	}
	config.AccessNo = conf.Get(srvName, confName, "accessNo").String("")
	if config.AccessNo == "" {
		return nil, errors.New(srvName+", accessNo is empty ")
	}
	config.SignType = conf.Get(srvName, confName, "signType").String("")
	if config.SignType == "" {
		return nil, errors.New(srvName+", signType is empty ")
	}
	config.UserId = conf.Get(srvName, confName, "userId").String("")
	if config.UserId == "" {
		return nil, errors.New(srvName+", userId is empty ")
	}
	config.StoreId = conf.Get(srvName, confName, "storeId").String("")
	if config.StoreId == "" {
		return nil, errors.New(srvName+", storeId is empty ")
	}
	config.TerminalId = conf.Get(srvName, confName, "terminalId").String("")
	if config.TerminalId == "" {
		return nil, errors.New(srvName+", terminalId is empty ")
	}
	config.MchNo = conf.Get(srvName, confName, "mchNo").String("")
	if config.MchNo == "" {
		return nil, errors.New(srvName+", mchNo is empty ")
	}
	config.AccessPrvKey = conf.Get(srvName, confName, "access-prv-key").String("")
	if config.AccessPrvKey == "" {
		return nil, errors.New(srvName+", access-prv-key is empty ")
	}
	config.AccessPubKey = conf.Get(srvName, confName, "access-pub-key").String("")
	if config.AccessPubKey == "" {
		return nil, errors.New(srvName+", access-pub-key is empty ")
	}
	config.PlatformPubKey = conf.Get(srvName, confName, "platform-pub-key").String("")
	if config.PlatformPubKey == "" {
		return nil, errors.New(srvName+", platform-pub-key is empty ")
	}

	isProduction := conf.Get(srvName, confName, "isProduction").Bool(false)

	bocpayClient, err = bocpay.New(config, isProduction)
	if err != nil {
		return nil, err
	}

	return bocpayClient, nil
}

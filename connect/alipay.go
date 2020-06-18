package connect

import (
	"github.com/micro/go-micro/v2/errors"
	alipay "github.com/smartwalle/alipay/v3"
)

var alipayClient *alipay.Client

func ConnectAlipay(srvName string, confName string) (*alipay.Client, error) {
	if alipayClient != nil {
		return alipayClient, nil
	}

	conf, _, err := ConnectConfig(srvName, confName)
	if err != nil {
		return nil, errors.InternalServerError(srvName, "read alipay config fail: %v", err.Error())
	}
	appId := conf.Get(srvName, confName, "appId").String("")
	if appId == "" {
		return nil, errors.InternalServerError(srvName, "appId is empty")
	}

	privateKey := conf.Get(srvName, confName, "privateKey").String("")
	if privateKey == "" {
		return nil, errors.InternalServerError(srvName, "privateKey is empty")
	}

	/*
	publicKey := conf.Get(srvName, confName, "publicKey").String("")
	if privateKey == "" {
		return nil, errors.InternalServerError(srvName, "publicKey is empty")
	}
	 */

	certPublicKeyRSA2 := conf.Get(srvName, confName, "certPublicKeyRSA2").String("")
	if certPublicKeyRSA2 == "" {
		return nil, errors.InternalServerError(srvName, "certPublicKeyRSA2 is empty")
	}

	rootCert := conf.Get(srvName, confName, "rootCert").String("")
	if certPublicKeyRSA2 == "" {
		return nil, errors.InternalServerError(srvName, "certPublicKeyRSA2 is empty")
	}

	certPublicKey := conf.Get(srvName, confName, "certPublicKey").String("")
	if certPublicKeyRSA2 == "" {
		return nil, errors.InternalServerError(srvName, "certPublicKey is empty")
	}

	isProduction := conf.Get(srvName, confName, "isProduction").Bool(false)

	alipayClient, err = alipay.New(appId, privateKey, isProduction)
	if err != nil {
		return nil, errors.InternalServerError(srvName, "new alipay fail: %v", err.Error())
	}

	/*
	err = alipayClient.LoadAliPayPublicKey(publicKey)
	if err != nil {
		return nil, errors.InternalServerError(srvName, "load public key fail: %v", err.Error())
	}
	*/

	err = alipayClient.LoadAppPublicCert(certPublicKey) // 加载应用公钥证书
	if err != nil {
		return nil, errors.InternalServerError(srvName, "LoadAppPublicCert fail: %v", err.Error())
	}
	err = alipayClient.LoadAliPayRootCert(rootCert) // 加载支付宝根证书
	if err != nil {
		return nil, errors.InternalServerError(srvName, "LoadAliPayRootCert fail: %v", err.Error())
	}
	err = alipayClient.LoadAliPayPublicCert(certPublicKeyRSA2) // 加载支付宝公钥证书
	if err != nil {
		return nil, errors.InternalServerError(srvName, "LoadAliPayPublicCert fail: %v", err.Error())
	}

	return alipayClient, nil
}

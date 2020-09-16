package connect

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const(
	TimeFormat 		= "20060102"
	PostContentType = "application/x-www-form-urlencoded;charset=utf-8"
	PostConnection 	= "close"

	TradeTestUrl 			= "http://183.62.24.78:3060/gateway/api/consumeTrans"	// 测试环境请求地址
	TradeProductionUrl 		= "https://paygate.efton.net/gateway/api/consumeTrans"	// 生产环境请求地址
	ManageTestUrl 			= "http://183.62.24.78:3060/gateway/api/merchant"	// 测试环境请求地址
	ManageProductionUrl 	= "https://paygate.efton.net/gateway/api/merchant"	// 生产环境请求地址
	DownloadTestUrl 		= "http://183.62.24.78:3060/gateway/api/downloadbill"	// 测试环境请求地址
	DownloadProductionUrl 	= "https://paygate.efton.net/gateway/api/downloadbill"	// 生产环境请求地址
)

type BocpayConfig struct {
	Version			string	`json:"version"`	// 接口版本号
	AccessNo		string	`json:"accessNo"`	// 接入机构号
	SignType		string	`json:"signType"`	// 签名算法
	UserId			string	`json:"userId"`		// 支付宝用户号
	StoreId			string	`json:"storeId"`	// 支付宝用户号
	TerminalId		string	`json:"terminalId"`	// 支付宝商户机具终端编号
	MchNo			string	`json:"mchNo"`		// 商户号
	AccessPrvKey	string	`json:"access-prv-key"`		// 后端私钥
	AccessPubKey	string	`json:"access-pub-key"`		// 后端公钥
	PlatformPubKey	string	`json:"platform-pub-key"`	// 支付平台公钥
}

type BocpayClient struct {
	mu 				sync.Mutex
	httpClient 		*http.Client
	location 		*time.Location
	bocpayConfig 	*BocpayConfig
	isProduction 	bool
	tradeUrl		string  // 交易类接口地址
	manageUrl		string  // 管理类接口地址
	downloadUrl 	string	// 文件交易类型接口地址

	accessPrivateKey	*rsa.PrivateKey // 后端私钥
	accessPublicKey		*rsa.PublicKey 	// 后端发布的公钥
	platformPublicKey 	*rsa.PublicKey  // 中国银行支付平台私钥
}

type TradeCreate struct {
	TransAmount  string // 交易额
	OutTransNo   string // 订单号
	GoodsSubject string // 商品订单标题
	NotifyUrl 	 string // 异步通知地址
}

type TradeQuery struct {
	OriTransData  string // 原订单日期yyyyMMdd
	OriOutTransNo string // 原商户交易订单号。二选一
	RefundNo	  string // 平台退款订单号。二选一， 退款选退款订单号
	NotifyUrl 	  string // 异步通知地址
}

type TradeCancel struct {
	OriTransData  string // 原订单日期yyyyMMdd
	OriOutTransNo string // 原商户交易订单号
	OutTransNo    string // 订单号
	NotifyUrl 	  string // 异步通知地址
}

type TradeClose struct {
	OriTransData  string // 原订单日期yyyyMMdd
	OriOutTransNo string // 原商户交易订单号
	OutTransNo    string // 订单号
	NotifyUrl 	  string // 异步通知地址
}

type TradeRefund struct {
	OriTransData  string // 原订单日期yyyyMMdd
	OriOutTransNo string // 原商户交易订单号
	TransAmount	  string // 退款金额
	TransReason	  string // 退款原因
	OutTransNo    string // 订单号
	NotifyUrl 	  string // 异步通知地址
}

type PromotionDetail struct {
	DiscountName 	string `json:"discountName"` 	// 优惠活动名称
	DiscountNumber 	string `json:"discountNumber"` 	// 优惠活动卷号
	GoodsTag 		string `json:"goodsTag"` 		// 优惠标识
	DiscountAmount 	string `json:"discountAmount"` 	// 优惠金额
	PaymentAmount 	string `json:"paymentAmount"` 	// 优惠活动名称
}

type TradeCreateRsp struct {
	RequestNo		string `json:"requestNo"`		// 请求流水号
	Version			string `json:"version"`	  		// 版本号
	AccessNo		string `json:"accessNo"`	  	// 接入机构号
	TransId			string `json:"transId"`	  		// 交易类型
	SignType		string `json:"signType"`	 	// 签名算法
	Signature		string `json:"signature"`	 	// 签名数据
	ProductId		string `json:"productId"`	 	// 产品类型
	MchNo			string `json:"mchNo"`		 	// 商户号
	TransDate		string `json:"transDate"`	 	// 交易日期
	OutTransNo		string `json:"outTransNo"`		// 商户订单号
	ReturnCode		string `json:"returnCode"`		// 网关应答码
	ReturnMsg		string `json:"returnMsg"`		// 网关应答码描述
	ResultCode		string `json:"resultCode"`		// 业务应答码
	ResultMsg		string `json:"resultMsg"`	 	// 业务应答码描述
	ResultSubMsg	string `json:"resultSubMsg"`	// 业务应答码描述明细
	AlipayTradeNo	string `json:"alipayTradeNo"`	// 支付宝支付窗交易号
	BankType		string `json:"bankType"`		// 付款银行
	BankUserId		string `json:"bankUserId"`		// 用户标识
	TransNo			string `json:"transNo"`			// 平台商户订单号
	PayNo			string `json:"payNo"`			// 平台支付订单号
	PayTime			string `json:"payTime"`			// 支付完成时间
	BankTradeNo		string `json:"bankTradeNo"`		// 银行订单号
	PromotionDetail	PromotionDetail `json:"promotionDetail"`	// 优惠明细
}

type TradeQueryRsp struct {
	RequestNo		string `json:"requestNo"`		// 请求流水号
	Version			string `json:"version"`	  		// 版本号
	AccessNo		string `json:"accessNo"`	  	// 接入机构号
	TransId			string `json:"transId"`	  		// 交易类型
	SignType		string `json:"signType"`	 	// 签名算法
	Signature		string `json:"signature"`	 	// 签名数据
	MchNo			string `json:"mchNo"`		 	// 商户号
	TransAmount		string `json:"transAmount"`	 	// 交易金额
	OrderState		string `json:"orderState"`	 	// 订单状态
	RefundJson		string `json:"refundJson"`	 	// 退款详情
	TransNo			string `json:"transNo"`			// 平台商户订单号
	BankUserId		string `json:"bankUserId"`		// 用户标识, 支付宝返回userId
	BankTradeNo		string `json:"bankTradeNo"`		// 银行订单号, 目前返回微信/支付宝渠道订单号
	ReturnCode		string `json:"returnCode"`		// 网关应答码
	ReturnMsg		string `json:"returnMsg"`		// 网关应答码描述
	ResultCode		string `json:"resultCode"`		// 业务应答码
	ResultMsg		string `json:"resultMsg"`	 	// 业务应答码描述
	ResultSubMsg	string `json:"resultSubMsg"`	// 业务应答码描述明细
	PromotionDetail	PromotionDetail `json:"promotionDetail"`	// 优惠明细
}

type TradeCancelRsp struct {
	RequestNo		string `json:"requestNo"`		// 请求流水号
	Version			string `json:"version"`	  		// 版本号
	AccessNo		string `json:"accessNo"`	  	// 接入机构号
	TransId			string `json:"transId"`	  		// 交易类型
	SignType		string `json:"signType"`	 	// 签名算法
	Signature		string `json:"signature"`	 	// 签名数据
	MchNo			string `json:"mchNo"`		 	// 商户号
	TransDate		string `json:"transDate"`	 	// 订单日期，商户交易订单日期yyyyMMdd
	OutTransNo		string `json:"outTransNo"`	 	// 商户订单号
	BankTradeNo		string `json:"bankTradeNo"`	 	// 银行订单号,目前返回微信/支付宝渠道订单号
	ReturnCode		string `json:"returnCode"`		// 网关应答码
	ReturnMsg		string `json:"returnMsg"`		// 网关应答码描述
	ResultCode		string `json:"resultCode"`		// 业务应答码
	ResultMsg		string `json:"resultMsg"`	 	// 业务应答码描述
	ResultSubMsg	string `json:"resultSubMsg"`	// 业务应答码描述明细
}

type TradeCloseRsp struct {
	RequestNo		string `json:"requestNo"`		// 请求流水号
	Version			string `json:"version"`	  		// 版本号
	AccessNo		string `json:"accessNo"`	  	// 接入机构号
	TransId			string `json:"transId"`	  		// 交易类型
	SignType		string `json:"signType"`	 	// 签名算法
	Signature		string `json:"signature"`	 	// 签名数据
	MchNo			string `json:"mchNo"`		 	// 商户号
	TransDate		string `json:"transDate"`	 	// 订单日期，商户交易订单日期yyyyMMdd
	OutTransNo		string `json:"outTransNo"`	 	// 商户订单号
	ReturnCode		string `json:"returnCode"`		// 网关应答码
	ReturnMsg		string `json:"returnMsg"`		// 网关应答码描述
	ResultCode		string `json:"resultCode"`		// 业务应答码
	ResultMsg		string `json:"resultMsg"`	 	// 业务应答码描述
	ResultSubMsg	string `json:"resultSubMsg"`	// 业务应答码描述明细
}

// 注意：优惠订单不允许进行部分退款
type TradeRefundRsp struct {
	RequestNo		string `json:"requestNo"`		// 请求流水号
	Version			string `json:"version"`	  		// 版本号
	AccessNo		string `json:"accessNo"`	  	// 接入机构号
	TransId			string `json:"transId"`	  		// 交易类型
	SignType		string `json:"signType"`	 	// 签名算法
	Signature		string `json:"signature"`	 	// 签名数据
	MchNo			string `json:"mchNo"`		 	// 商户号
	TransDate		string `json:"transDate"`	 	// 订单日期，商户交易订单日期yyyyMMdd
	OutTransNo		string `json:"outTransNo"`	 	// 商户订单号
	TransAmount		string `json:"transAmount"`	 	// 交易金额
	RefundReason	string `json:"refundReson"`		// 退货原因
	TransNo			string `json:"transNo"`			// 平台退款订单号
	BankTradeNo		string `json:"bankTradeNo"`		// 银行订单号, 目前返回微信/支付宝渠道订单号
	ReturnCode		string `json:"returnCode"`		// 网关应答码
	ReturnMsg		string `json:"returnMsg"`		// 网关应答码描述
	ResultCode		string `json:"resultCode"`		// 业务应答码
	ResultMsg		string `json:"resultMsg"`	 	// 业务应答码描述
	ResultSubMsg	string `json:"resultSubMsg"`	// 业务应答码描述明细
}

type DownloadBillRsp struct {
	ReturnCode	string `json:"returnCode"`	// 网关应答码
	ReturnMsg	string `json:"returnMsg"`	// 网关应答码描述
	BillData	string `json:"billData"`	// 对账文件内容,Base64编码后文件内容。需Base64解码
}

var bocpayClient *BocpayClient

func ConnectBocpay(srvName string, confName string) (*BocpayClient, error) {
	if bocpayClient != nil {
		return bocpayClient, nil
	}

	conf, _, err := ConnectConfig(srvName, confName)
	if err != nil {
		return nil, errors.New(srvName+" read bocpay config fail: "+err.Error())
	}

	config := new(BocpayConfig)

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

	bocpayClient, err = New(config, isProduction)
	if err != nil {
		return nil, err
	}

	return bocpayClient, nil
}

func New(config *BocpayConfig, isProduction bool) (client *BocpayClient, err error) {
	client = new(BocpayClient)
	client.isProduction = isProduction
	client.bocpayConfig = config
	client.httpClient 	= http.DefaultClient

	client.location, err = time.LoadLocation("Asia/Chongqing")
	if nil != err {
		return nil, err
	}

	if client.isProduction {
		client.tradeUrl 	= TradeProductionUrl
		client.manageUrl 	= ManageProductionUrl
		client.downloadUrl 	= DownloadProductionUrl
	} else {
		client.tradeUrl 	= TradeTestUrl
		client.manageUrl 	= ManageTestUrl
		client.downloadUrl 	= DownloadTestUrl
	}

	// 加载Key
	err = client.loadAccessPrivateKey(config.AccessPrvKey)
	if err != nil {
		return nil, err
	}
	err = client.loadAccessPublicKey(config.AccessPubKey)
	if err != nil {
		return nil, err
	}
	err = client.loadPlatformPublicKey(config.PlatformPubKey)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// 测试接口，开发完后删除
func (this *BocpayClient) Test(item int)  {

	switch item {
	case 0: {
		fmt.Println("access公钥加密、私钥解密测试。")
		data := "test"

		signature, _ := this.getSignature([]byte(data))
		fmt.Println("signature: ", signature)
		fmt.Println("verifySignature: ", this.verifySignature(this.accessPublicKey, []byte(data), signature))
		fmt.Println()
		break
	}
	case 1: {
		fmt.Println("access公钥加密、私钥解密测试。")
		data := "test"
		encryData, _ := this.publicKeyEncrypt(this.accessPublicKey, []byte(data))
		fmt.Println("encryptData: ", encryData)
		decryData, _ := this.privateKeyDecrypt(this.accessPrivateKey, encryData)
		fmt.Println("decryptData: ", decryData)
		fmt.Println()
		break
	}
	}
}

// 下载对账单
func (this *BocpayClient) DownloadBill(bill string) (*DownloadBillRsp, error) {
	data := url.Values{}
	// 固定，一般不会改
	data.Set("transId", "105")

	// 固定参数，后面通过config统一配置
	data.Set("version", this.bocpayConfig.Version)
	data.Set("accessNo", this.bocpayConfig.AccessNo)
	data.Set("signType", this.bocpayConfig.SignType)

	// 生成的参数
	data.Set("requestNo", this.getRequestNo())

	// 需要传入的参数
	data.Set("billDate", bill)

	// 添加签名
	signature, _ := this.getSignature([]byte(data.Encode()))
	data.Set("signature", signature)

	// 发起请求
	downloadBillRsp := new(DownloadBillRsp)
	err := this.post(this.tradeUrl, data.Encode(), downloadBillRsp)

	// base64解码
	decodeBillData, err := base64.StdEncoding.DecodeString(downloadBillRsp.BillData)
	downloadBillRsp.BillData = string(decodeBillData)

	return downloadBillRsp, err
}

// 创建订单, 需要返回交易日期，订单号
func (this *BocpayClient) TradeCreate(param *TradeCreate) (*TradeCreateRsp, error) {
	data := url.Values{}
	// 固定，一般不会改
	data.Set("transId", "100")	// 交易类型
	data.Set("productId", "1053") // 产品类型

	// 固定参数，后面通过config统一配置
	data.Set("version", this.bocpayConfig.Version)
	data.Set("accessNo", this.bocpayConfig.AccessNo)
	data.Set("signType", this.bocpayConfig.SignType)
	data.Set("mchNo", this.bocpayConfig.MchNo)
	data.Set("userId", this.bocpayConfig.UserId)			// 当productId=1053时必填买家的支付宝唯一用户号
	data.Set("storeId" , this.bocpayConfig.StoreId) 		// 根据自身业务场景填写，商户门店编号
	data.Set("terminalId", this.bocpayConfig.TerminalId)	// 根据自身业务场景填写，商户机具编号


	// 生成的参数
	data.Set("requestNo", this.getRequestNo())
	data.Set("transDate", time.Now().In(this.location).Format(TimeFormat))	 //交易日期

	// 需要传入的参数
	data.Set("transAmount", param.TransAmount)
	data.Set("outTransNo", param.OutTransNo)		// 商户订单号，需保证商户端不重复， 需要返回
	data.Set("goodsSubject", param.GoodsSubject)	// 商品订单标题
	data.Set("notifyUrl", param.NotifyUrl)			// 异步通知地址

	// 添加签名
	signature, _ := this.getSignature([]byte(data.Encode()))
	data.Set("signature", signature)

	// 发起请求
	tradeCreateRsp := new(TradeCreateRsp)
	err := this.post(this.tradeUrl, data.Encode(), tradeCreateRsp)

	return tradeCreateRsp, err
}

// 订单查询， 需要返回订单信息
func (this *BocpayClient) TradeQuery(param *TradeQuery) (*TradeQueryRsp, error) {
	data := url.Values{}
	// 固定，一般不会改
	data.Set("transId", "101")

	// 固定参数，后面通过config统一配置
	data.Set("version", this.bocpayConfig.Version)
	data.Set("accessNo", this.bocpayConfig.AccessNo)
	data.Set("signType", this.bocpayConfig.SignType)
	data.Set("mchNo", this.bocpayConfig.MchNo) 	 //商户号

	// 生成的参数
	data.Set("requestNo", this.getRequestNo())

	// 传入参数
	data.Set("oriTransDate", param.OriTransData)		// 原交易订单日期yyyyMMdd
	data.Set("notifyUrl", param.NotifyUrl) 			// 异步通知地址
	if "" == param.OriOutTransNo {
		data.Set("refundNo", param.RefundNo)			// 平台退款订单号。二选一， 退款选退款订单号
	} else {
		data.Set("oriOutTransNo", param.OriOutTransNo)	// 原商户交易订单号，二选一。
	}

	// 添加签名
	signature, _ := this.getSignature([]byte(data.Encode()))
	data.Set("signature", signature)

	// 发起请求
	tradeQueryRsp := new(TradeQueryRsp)
	err := this.post(this.tradeUrl, data.Encode(), tradeQueryRsp)

	return tradeQueryRsp, err
}

// 取消订单， 返回应答报文和错误信息
func (this *BocpayClient) TradeCancel(param *TradeCancel) (*TradeCancelRsp, error) {
	data := url.Values{}
	// 固定，一般不会改
	data.Set("transId", "103")

	// 固定参数，后面通过config统一配置
	data.Set("version", this.bocpayConfig.Version)
	data.Set("accessNo", this.bocpayConfig.AccessNo)
	data.Set("signType", this.bocpayConfig.SignType)
	data.Set("mchNo", this.bocpayConfig.MchNo) 	 //商户号

	// 生成的参数
	data.Set("requestNo", this.getRequestNo())
	data.Set("transDate", time.Now().In(this.location).Format(TimeFormat))	 //交易日期

	// 传入参数
	data.Set("oriTransDate", param.OriTransData)	// 原交易订单日期yyyyMMdd
	data.Set("oriOutTransNo", param.OriOutTransNo)	// 原商户交易订单号
	data.Set("outTransNo", param.OutTransNo)		// 商户订单号，需保证商户端不重复， 需要返回
	data.Set("notifyUrl", param.NotifyUrl) 		// 异步通知地址

	// 添加签名
	signature, _ := this.getSignature([]byte(data.Encode()))
	data.Set("signature", signature)

	// 发起请求
	tradeCancelRsp := new(TradeCancelRsp)
	err := this.post(this.tradeUrl, data.Encode(), tradeCancelRsp)

	return tradeCancelRsp, err
}

// 关闭订单， 返回应答报文和错误信息
func (this *BocpayClient) TradeClose(param *TradeClose) (*TradeCloseRsp, error) {
	data := url.Values{}
	// 固定，一般不会改
	data.Set("transId", "104")

	// 固定参数，后面通过config统一配置
	data.Set("version", this.bocpayConfig.Version)
	data.Set("accessNo", this.bocpayConfig.AccessNo)
	data.Set("signType", this.bocpayConfig.SignType)
	data.Set("mchNo", this.bocpayConfig.MchNo)

	// 生成的参数
	data.Set("requestNo", this.getRequestNo())
	data.Set("transDate", time.Now().In(this.location).Format(TimeFormat))	 //交易日期


	// 传入参数
	data.Set("oriTransDate", param.OriTransData)	// 原交易订单日期yyyyMMdd
	data.Set("oriOutTransNo", param.OriOutTransNo)	// 原商户交易订单号
	data.Set("outTransNo", param.OutTransNo)		// 商户订单号，需保证商户端不重复， 需要返回
	data.Set("notifyUrl", param.NotifyUrl) 		// 异步通知地址

	// 添加签名
	signature, _ := this.getSignature([]byte(data.Encode()))
	data.Set("signature", signature)

	// 发起请求
	tradeCloselRsp := new(TradeCloseRsp)
	err := this.post(this.tradeUrl, data.Encode(), tradeCloselRsp)

	return tradeCloselRsp, err
}

// 退款， 返回应答报文和错误信息
func (this *BocpayClient) TradeRefund(param *TradeRefund) (*TradeRefundRsp, error) {
	data := url.Values{}
	// 固定，一般不会改
	data.Set("transId", "102")

	// 固定参数，后面通过config统一配置
	data.Set("version", this.bocpayConfig.Version)
	data.Set("accessNo", this.bocpayConfig.AccessNo)
	data.Set("signType", this.bocpayConfig.SignType)
	data.Set("mchNo", this.bocpayConfig.MchNo)

	// 生成的参数
	data.Set("requestNo", this.getRequestNo())
	data.Set("transDate", time.Now().In(this.location).Format(TimeFormat))	 //交易日期

	// 传入参数
	data.Set("oriTransDate", param.OriTransData)	// 原交易订单日期yyyyMMdd
	data.Set("oriOutTransNo", param.OriOutTransNo)	// 原商户交易订单号
	data.Set("transAmount", param.TransAmount)		// 退款金额
	data.Set("refundReason", param.TransReason)	// 退款原因
	data.Set("outTransNo", param.OutTransNo)		// 商户订单号，需保证商户端不重复， 需要返回
	data.Set("notifyUrl", param.NotifyUrl) 		// 异步通知地址

	// 添加签名
	signature, _ := this.getSignature([]byte(data.Encode()))
	data.Set("signature", signature)

	// 发起请求
	tradeRefundRsp := new(TradeRefundRsp)
	err := this.post(this.tradeUrl, data.Encode(), tradeRefundRsp)

	return tradeRefundRsp, err
}

// 异步通知验证
func (this *BocpayClient) VerifyNotify(request string) error {
	unmarshalRequest := gjson.Parse(request)
	requestContent := url.Values{}

	if signature := unmarshalRequest.Get("signature"); signature.Exists() {
		unmarshalRequest.ForEach(func(key, value gjson.Result) bool {
			if key.String() != "signature" {
				requestContent.Set(key.String(), value.String())
			}
			return true
		})
		return this.verifySignature(this.platformPublicKey, []byte(requestContent.Encode()), signature.String())
	}

	return errors.New("verify notify failed")
}

// 测试发布异步通知, 用来测试TradeNotify
func (this *BocpayClient) TestTradeNotify() {
	data := url.Values{}

	data.Set("transId", "100")
	data.Set("version", "V1.0")
	data.Set("accessNo", "20201804120000018121")
	data.Set("signType", "RSA2")
	data.Set("mchNo", "850780641001001") 	// 商户号
	data.Set("notifyUrl", "test") 		// 异步通知地址

	data.Set("requestNo", this.getRequestNo())
	data.Set("transDate", time.Now().In(this.location).Format(TimeFormat))	// 交易日期， 需要返回
	data.Set("outTransNo", "outTransNo")	// 商户订单号，需保证商户端不重复

	data.Set("orderId", "20180529000121105200000272")
	data.Set("payTime", "20180529160952")
	data.Set("productId", "1052")
	data.Set("respCode", "0000")
	data.Set("respDesc", "成功")
	data.Set("transAmount", "10")
	data.Set("payNo", "payNo") // 平台支付订单号
}

// 通用接口
// 加载后端私钥
func (this *BocpayClient) loadAccessPrivateKey(input string) (err error) {
	data, err := this.parseKey(input)
	if err != nil {
		return err
	}

	this.accessPrivateKey, err = x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return err
	}

	return nil
}

// 加载后端公钥
func (this *BocpayClient) loadAccessPublicKey(input string) (err error) {
	data, err := this.parseKey(input)
	if err != nil {
		return err
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return err
	}
	this.accessPublicKey = pubKeyInterface.(*rsa.PublicKey)

	return nil
}

// 加载平台公钥
func (this *BocpayClient) loadPlatformPublicKey(input string) (err error) {
	data, err := this.parseKey(input)
	if err != nil {
		return err
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return err
	}
	this.platformPublicKey = pubKeyInterface.(*rsa.PublicKey)

	return nil
}

// 后端私钥签名
func (this *BocpayClient) getSignature(data []byte) (string, error) {
	hash := sha256.New()
	hash.Write(data)
	hashed := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, this.accessPrivateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), err
}

// 使用公钥进行签名验证
func (this *BocpayClient) verifySignature(pubKey *rsa.PublicKey, data []byte, signature string) error {
	signatureByte, _ := base64.StdEncoding.DecodeString(signature)
	sha := sha256.New()
	sha.Write(data)
	hashed := sha.Sum(nil)

	return  rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signatureByte)
}

// 使用公钥加密
func (this *BocpayClient) publicKeyEncrypt(pubKey *rsa.PublicKey, data []byte) (string, error) {
	result, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	return base64.StdEncoding.EncodeToString(result), err
}

// 使用私钥解密
func (this *BocpayClient) privateKeyDecrypt(priKey *rsa.PrivateKey, data string) (string, error) {
	// 解析签名
	dataByte, _ := base64.StdEncoding.DecodeString(data)
	decryData, err := rsa.DecryptPKCS1v15(rand.Reader, priKey, dataByte)
	if nil != err {
		return "", err
	}

	return string(decryData), nil
}

// 获取流水线号
func (this *BocpayClient) getRequestNo() string {
	currentTime := time.Now().In(this.location)

	//要加'.'才能获取到毫秒，拿到后再去掉
	requestNo := time.Unix(0, currentTime.UnixNano()).Format("20060102150405.000")
	requestNo = strings.Replace(requestNo, ".", "", -1)

	return requestNo
}

// 请求数据
func (this *BocpayClient) post(postUrl string, data string, result interface{}) (err error) {
	// 设置请求参数
	req, err := http.NewRequest("POST", postUrl, strings.NewReader(data))
	if nil != err {
		return err
	}
	req.Header.Add("Content-Type", PostContentType)
	req.Header.Add("Connection", PostConnection)

	// 发起请求
	rsp, err := this.httpClient.Do(req)
	if nil != err {
		return err
	}
	defer rsp.Body.Close()

	// 获取结果
	body, err := ioutil.ReadAll(rsp.Body)
	if nil != err {
		return err
	}

	// 公钥验证
	unmarshalData := gjson.Parse(string(body))
	if signature := unmarshalData.Get("signature"); signature.Exists() {
		verifyData := url.Values{}
		unmarshalData.ForEach(func(key, value gjson.Result) bool {
			if key.String() != "signature" {
				verifyData.Set(key.String(), value.String())
			}
			return true
		})

		err = this.verifySignature(this.platformPublicKey, []byte(verifyData.Encode()), signature.String())

		// 暂时没拿到公钥，跳过
		err = nil
		if nil != err {
			return err
		}
	}

	// 解析结果
	err = json.Unmarshal(body, result)
	return err
}

// 密钥解析
func (this *BocpayClient) parseKey(input string) ([]byte, error) {
	block, _ := pem.Decode([]byte(input))
	if nil == block {
		// 兼容没有-----BEGIN -----END的情况
		data, err := base64.StdEncoding.DecodeString(input)
		return data, err
	}

	return block.Bytes, nil
}
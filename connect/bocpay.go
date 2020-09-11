package connect

import (
	"crypto"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type BocpayClient struct {
	mu 			sync.Mutex
	Client 		*http.Client
	location 	*time.Location

	accessPrivateKey	*rsa.PrivateKey // 后端私钥
	accessPublicKey		*rsa.PublicKey 	// 后端发布的公钥
	platformPublicKey 	*rsa.PublicKey  // 中国银行支付平台私钥
}

type BocpayCardInfo struct {
	CardNo		string	`json:"cardNo"`
	CardHolder	string	`json:"cardHolder"`
	CardId		string	`json:"cardId"`
	CardPhone	string	`json:"cardPhone"`
	CardPeriod	string	`json:"cardPeriod"`
	CardCVN2	string	`json:"cardCVN2"`
}

type BocpayAddress struct {
	Longitude 	string	`json:"longitude"`
	Latitude 	string	`json:"latitude"`
}

type BocpayQuickPay struct {
	TransAmount string
	CardInfo 	*BocpayCardInfo
	Address 	*BocpayAddress
}

func ConnectBocpay(srvName string, confName string) (*BocpayClient, error) {
	fmt.Println(srvName, confName)
	client := new(BocpayClient)

	accessPrivateCertify, err := getCertifyData("/Users/wxx/wxx_nice/dev_test/key/access-prv-key.pem")
	if err != nil {
		return nil, err
	}
	accessPublicCertify, err := getCertifyData("/Users/wxx/wxx_nice/dev_test/key/access-pub-key.pem")
	if err != nil {
		return nil, err
	}
	platformPublicCertify, err := getCertifyData("/Users/wxx/wxx_nice/dev_test/key/platform-pub-key.pem")
	if err != nil {
		return nil, err
	}

	err = client.LoadAccessPrivateKey(accessPrivateCertify)
	if err != nil {
		return nil, err
	}
	err = client.LoadAccessPublicKey(accessPublicCertify)
	if err != nil {
		return nil, err
	}
	err = client.LoadPlatformPublicKey(platformPublicCertify)
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
	case 2: {
		fmt.Println("aecEncrypt加密、解密。")
		key, _ := hex.DecodeString("50505F7C20251286AA92A501BC0415E4")
		encry := this.aecEncryptECB([]byte("test"), key)
		fmt.Println("aecEncryptECB encry: ", hex.EncodeToString(encry))
		decry := this.aecDecryptECB(encry, key)
		fmt.Println("aecEncryptECB decry: ", string(decry))

		demoStr := "232eb3893aee65e1ee99c3fc3c36d061b0b7c62cae1dd3e250e69ba0183bf4f1343a32e8ab687031b910f43d2f027eb96fdd2138c603ea5505905af71a21438fd2323609b18956ff781de9b2e9b58623a5b3f7bde46f71cf1d2639f6a161d9ee0a41b3477e6c3b8ae171f3bbc885fa4796404856406702ad665aecab54f079d202ec913fea3a1bc310f48a3153417be287a2ce96fff50b14111e0e278361de08"
		demoByte, _ := hex.DecodeString(demoStr)
		decry = this.aecDecryptECB(demoByte, key)
		fmt.Println("aecEncryptECB demo decry: ", string(decry))
		fmt.Println()
		break
	}
		
	}
	/*
	// 验证支付平台的签名， 公钥不对
	data := "returnCode=0045&returnMsg=接入机构号错误"
	signature := "cNHIgSvUgqWWt3OJEM//EohxAP8gG5qrkCy6Iz+eUJS+NQo/N5dQiBmofSWyPyttns8Ysz/lLo7OOrmQmmUpC7zvB0Hde2EOcPrXaWvD0s+/FEIZvi89r0esG8iaopD6A+g2axfGioVbAkhCd57m/aj5rHu6ld/58UivozuTpHwHSs8xuU6HQ/yDJJU9GmUD8IoLrDYBKZZRVRSQzTXtzwkpKmlWOXZ0vgxmc8fXiIBrGfkkAeTCp94Dasx2jhWaKXVQrX9wwzdtwMZTzigFQsFIPaFWOv8zduImJMfBK7N1EpMR9o9zFkOdbxGbPTiMndLXMuIETUa8h83RWfDs3Q=="
	pubKey, _ := getPlatformPublicKey()
	fmt.Println(verifySignature(pubKey, []byte(data), signature))
	//*/

}

func (this *BocpayClient) New() {

}

// 下载对账单
func (this *BocpayClient) DownloadBill(bill string) {
	data := url.Values{}
	data.Set("requestNo", this.getRequestNo())
	data.Set("version", "V1.0")
	data.Set("transId", "105")
	data.Set("accessNo", "20201804120000018121")
	data.Set("signType", "RSA2")
	data.Set("billDate", bill)
	signature, _ := this.getSignature([]byte(data.Encode()))
	data.Set("signature", signature)

	postUrl := "http://183.62.24.78:3060/gateway/api/downloadbill"

	resultStr, _ := this.send(postUrl, data.Encode())

	result := make(map[string]string, 10)
	json.Unmarshal([]byte(resultStr), &result)
	resSignature := ""
	if value, ok := result["signature"]; ok {
		resSignature = value
		delete(result, "signature")
	}
	fmt.Println("result: ", result)

	res := url.Values{}
	for key, value := range result {
		res.Set(key, value)
	}

	err := this.verifySignature(this.platformPublicKey, []byte(res.Encode()), resSignature)

	if nil == err {
		fmt.Println("success.")
	}
}

// 快捷支付
func (this *BocpayClient) QuickPay(quickPay *BocpayQuickPay) error {
	aecKey, _ := hex.DecodeString("50505F7C20251286AA92A501BC0415E4")

	cardInfo, err := json.Marshal(quickPay.CardInfo)
	if nil != err {
		fmt.Println("Marshal cardInfo failed")
		return err
	}

	address, err := json.Marshal(quickPay.Address)
	if nil != err {
		fmt.Println("Marshal address failed")
		return err
	}

	data := url.Values{}
	data.Set("requestNo", this.getRequestNo())
	data.Set("version", "V1.0")
	data.Set("transId", "106")
	data.Set("accessNo", "20201804120000018121")
	data.Set("signType", "RSA2")

	data.Set("productId", "1101") 		 //产品类型
	data.Set("mchNo", "850780641001001") 	 //商户号
	data.Set("transDate", "格式yyyyMMdd")	 //交易日期
	data.Set("outTransNo", "RSA2")		 //商户订单号，需保证商户端不重复
	data.Set("goodsSubject", "测试交易商品") //商品订单标题

	data.Set("notifyUrl", "后面加") 	// 异步通知地址
	data.Set("webNotifyUrl", "后面加") //页面通知地址

	data.Set("transAmount", quickPay.TransAmount)
	data.Set("cardInfo", hex.EncodeToString(this.aecEncryptECB(cardInfo, aecKey)))
	data.Set("address", string(address))

	fmt.Println("data: ", data)
	postUrl := "http://183.62.24.78:3060/gateway/api/consumeTrans"
	_, err = this.send(data.Encode(), postUrl)
	return err
}

// 加载后端似钥
func (this *BocpayClient) LoadAccessPrivateKey(data []byte) error {
	var err error
	this.accessPrivateKey, err = x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return err
	}

	return nil
}

// 加载后端公钥
func (this *BocpayClient) LoadAccessPublicKey(data []byte) error {
	pubKeyInterface, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return err
	}
	this.accessPublicKey = pubKeyInterface.(*rsa.PublicKey)

	return nil
}

// 加载平台公钥
func (this *BocpayClient) LoadPlatformPublicKey(data []byte) error {
	pubKeyInterface, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return err
	}
	this.platformPublicKey = pubKeyInterface.(*rsa.PublicKey)

	return nil
}

// 通用接口
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
	loc, _ := time.LoadLocation("Asia/Shanghai")
	currentTime := time.Now().In(loc)

	//要加'.'才能获取到毫秒，拿到后再去掉
	requestNo := time.Unix(0, currentTime.UnixNano()).Format("20060102150405.000")
	requestNo = strings.Replace(requestNo, ".", "", -1)

	return requestNo
}

// 请求数据
func (this *BocpayClient) send(postUrl string, data string) (string, error) {
	client := &http.Client{}
	// 设置请求参数
	req, err := http.NewRequest("POST", postUrl, strings.NewReader(data))
	if nil != err {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;;charset=utf-8")
	req.Header.Add("Connection", "close")

	// 发起请求
	rsp, err := client.Do(req)
	if nil != err {
		return "", err
	}
	defer rsp.Body.Close()

	// 获取结果
	body, err := ioutil.ReadAll(rsp.Body)
	if nil != err {
		return "", err
	}
	return string(body), nil
}

// aec加密信息， 比如卡信息
func (this *BocpayClient) aecEncryptECB(data, key []byte) []byte {
	cipher, _ := aes.NewCipher(key)
	length := (len(data) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, data)
	pad := byte(len(plain) - len(data))
	for i := len(data); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted := make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, cipher.BlockSize(); bs <= len(data); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted
}

// aec解密信息， 比如卡信息
func (this *BocpayClient) aecDecryptECB(data, key []byte) []byte {
	cipher, _ := aes.NewCipher(key)
	decrypted := make([]byte, len(data))
	//
	for bs, be := 0, cipher.BlockSize(); bs < len(data); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], data[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim]
}


// 临时读取，后期换成微服务config
func getCertifyData(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("get privatekey failed!")
	}

	block, _ := pem.Decode(data)
	if nil == block {
		return base64.StdEncoding.DecodeString(string(data))
	}

	return block.Bytes, nil
}

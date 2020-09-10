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
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type BocpayClient struct {
	mu 			sync.Mutex
	Client 		*http.Client
	location 	*time.Location

	appPrivateKey		*rsa.PrivateKey 	// 后端私钥
	platformPublicKey 	*rsa.PublicKey // 中国银行支付平台私钥
}

func ConnectBocpay(srvName string, confName string) (*BocpayClient) {
	fmt.Println(srvName, confName)
	return new(BocpayClient)
}

func (this *BocpayClient) Test()  {
	/*
	// 验证支付平台的签名， 公钥不对
	data := "returnCode=0045&returnMsg=接入机构号错误"
	signature := "cNHIgSvUgqWWt3OJEM//EohxAP8gG5qrkCy6Iz+eUJS+NQo/N5dQiBmofSWyPyttns8Ysz/lLo7OOrmQmmUpC7zvB0Hde2EOcPrXaWvD0s+/FEIZvi89r0esG8iaopD6A+g2axfGioVbAkhCd57m/aj5rHu6ld/58UivozuTpHwHSs8xuU6HQ/yDJJU9GmUD8IoLrDYBKZZRVRSQzTXtzwkpKmlWOXZ0vgxmc8fXiIBrGfkkAeTCp94Dasx2jhWaKXVQrX9wwzdtwMZTzigFQsFIPaFWOv8zduImJMfBK7N1EpMR9o9zFkOdbxGbPTiMndLXMuIETUa8h83RWfDs3Q=="
	pubKey, _ := getPlatformPublicKey()
	fmt.Println(verifySignature(pubKey, []byte(data), signature))
	//*/

	/*
	// access公钥加密、私钥解密测试。
	data := "test"
	priKey, _ := getAccessPrivateKey()
	pubKey, _ := getAccessPublicKey()

	encryData, _ := publicKeyEncrypt(pubKey, []byte(data))
	fmt.Println("encryptData: ", encryData)
	decryData, _ := privateKeyDecrypt(priKey, encryData)
	fmt.Println("decryptData: ", decryData)
	 //*/

	/*
	// 私钥签名，公钥验证
	data := "test"
	priKey, _ := getAccessPrivateKey()
	pubKey, _ := getAccessPublicKey()

	signature, _ := getSignature(priKey, []byte(data))
	fmt.Println("signature: ", signature)
	fmt.Println("verifySignature: ", verifySignature(pubKey, []byte(data), signature))
	//*/
}

func (this *BocpayClient) DownloadBill(bill string) {
	priKey,_ := getAccessPrivateKey()

	data := url.Values{}
	data.Set("requestNo", getRequestNo())
	data.Set("version", "V1.0")
	data.Set("transId", "105")
	data.Set("accessNo", "20201804120000018121")
	data.Set("signType", "RSA2")
	data.Set("billDate", bill)
	signature, _ := getSignature(priKey, []byte(data.Encode()))
	data.Set("signature", signature)

	postUrl := "http://183.62.24.78:3060/gateway/api/downloadbill"

	resultStr, _ := send(postUrl, data.Encode())

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

	pubKey, _ := getPlatformPublicKey()
	err := verifySignature(pubKey, []byte(res.Encode()), resSignature)

	if nil == err {
		fmt.Println("success.")
	}
}

// 请求数据
func send(postUrl string, data string) (string, error) {
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

// 获取流水号
func getRequestNo() string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	currentTime := time.Now().In(loc)

	//要加'.'才能获取到毫秒，拿到后再去掉
	requestNo := time.Unix(0, currentTime.UnixNano()).Format("20060102150405.000")
	requestNo = strings.Replace(requestNo, ".", "", -1)

	return requestNo
}

// 将请求数据格式整理转换
func getPostStr(post map[string]string) (postStr string) {
	postStr = ""

	if 0 != len(post) {
		//对key进行排序
		var keys []string
		for key, _ := range post {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			postStr += key + "=" + post[key] + "?"
		}

		//去掉末尾对'?'
		postStr = postStr[:len(postStr) - 1]
	}

	return postStr
}

// 私钥加密， 获取签名。
func getSignature(certPrivateKey []byte, data []byte) (string, error) {
	pri, err := x509.ParsePKCS1PrivateKey(certPrivateKey)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(data)
	hashed := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, pri, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), err
}

// 公钥加密
func publicKeyEncrypt(publicKey []byte, data []byte) (string, error) {
	pubInterface, _ := x509.ParsePKIXPublicKey(publicKey)
	pubKey := pubInterface.(*rsa.PublicKey)
	result, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)

	return base64.StdEncoding.EncodeToString(result), err
}

// Access私钥获取, 自己生成的私钥。
func getAccessPrivateKey() ([]byte, error) {
	file := "/Users/wxx/wxx_nice/dev_test/key/access-prv-key.pem"
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("get privatekey failed!")
	}

	block, _ := pem.Decode(data)
	if nil == block || block.Type != "RSA PRIVATE KEY"{
		fmt.Println("faile get block.")
		//er = ""
	}

	return block.Bytes, err
}

// Access公钥获取, 自己生成的公钥。
func getAccessPublicKey() ([]byte, error)  {
	file := "/Users/wxx/wxx_nice/dev_test/key/access-pub-key.pem"
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("get privatekey failed!")
	}

	block, _ := pem.Decode(data)
	if nil == block || block.Type != "PUBLIC KEY"{
		fmt.Println("faile get block.")
		//er = ""
	}

	return block.Bytes, err
}

// Platform公钥获取，支付平台给的公钥。
func getPlatformPublicKey() ([]byte, error) {
	file := "/Users/wxx/wxx_nice/dev_test/key/platform-pub-key.pem"
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("get getPlatformPublicKey failed!")
		return nil, nil
	}

	fmt.Println("pub: ", string(data))
	return base64.StdEncoding.DecodeString(string(data))
}

// 使用公钥进行签名验证
func verifySignature(pubKey []byte, data []byte, signature string) (error) {
	signatureByte, _ := base64.StdEncoding.DecodeString(signature)

	pubInterface, _ := x509.ParsePKIXPublicKey(pubKey)
	pub := pubInterface.(*rsa.PublicKey)

	sha := sha256.New()
	sha.Write(data)
	hashed := sha.Sum(nil)

	return  rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], signatureByte)
}

// 获取私钥解析数据
func privateKeyDecrypt(priKey []byte, data string) (string, error) {
	// 私钥解析
	priv, err := x509.ParsePKCS1PrivateKey(priKey)
	if err != nil {
		return "", err
	}

	// 解析签名
	dataByte, _ := base64.StdEncoding.DecodeString(data)
	decryData, err := rsa.DecryptPKCS1v15(rand.Reader, priv, dataByte)
	if nil != err {
		return "", err
	}

	return string(decryData), nil
}
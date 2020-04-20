package library

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

var httpTimeOut = 15 * time.Second

/*
http-get请求
*/
func HttpGet(url string) (string, error) {

	// 超时时间：5秒
	client := &http.Client{Timeout: httpTimeOut}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return string(result), err
}

/*
http-post请求
*/
func HttpPost(url string, data []byte, contentType string) (string, error) {
	// 超时时间：5秒
	client := &http.Client{Timeout: httpTimeOut}
	resp, err := client.Post(url, contentType, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return string(result), err
}

package library

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 2 * time.Second,
}

func HttpGet(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return string(result), err
}

func HttpPost(url string, data []byte, contentType string) (string, error) {
	resp, err := httpClient.Post(url, contentType, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return string(result), err
}

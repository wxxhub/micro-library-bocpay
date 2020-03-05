package helper

import (
	"net/http"
	_ "net/http/pprof"
)

func PprofHttpListen() {
	go func() {
		ip := "0.0.0.0:6060"
		http.ListenAndServe(ip, nil)
	}()
}

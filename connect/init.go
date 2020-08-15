package connect

import (
	"github.com/opentracing/opentracing-go"
	"log"
)

func Initialization(serviceName string) opentracing.Tracer {
	err := ConnectLog(serviceName)
	if err != nil {
		log.Fatal(err)
	}

	err = ConnectStdLog(serviceName)
	if err != nil {
		log.Fatal(err)
	}

	tracer, err := ConnectJaeger(serviceName)
	if err != nil {
		log.Println("connect jaeger err: ", err)
	}

	return tracer
}

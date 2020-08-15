package connect

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"time"
)

func ConnectJaeger(srvName string) (opentracing.Tracer, error) {
	sender, err := jaeger.NewUDPTransport("jaeger-agent.default:5775", 0)
	if err != nil {
		return nil, err
	}

	tracer, _ := jaeger.NewTracer(
		srvName,
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(
			sender,
			jaeger.ReporterOptions.BufferFlushInterval(1*time.Second)),
	)
	return tracer, nil
}

package connect

import (
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var writers sync.Map

type kafkaOption struct {
	Broker []string
}

type connection struct {
	w      *kafka.Writer
	broker []string
	once   sync.Once
}

func GetKafkaWriter(topic string, log *logrus.Entry) (w *kafka.Writer, err error) {
	c, ok := writers.Load(topic)
	if !ok {
		conf, watcher, err := ConnectConfig("kafka", "broker")
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("kafka connect config")
			return nil, fmt.Errorf("kafka connect config: %w", err)
		}

		var option = new(kafkaOption)
		err = conf.Get("kafka", "broker").Scan(option)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("config scan kafka")
			return nil, fmt.Errorf("config scan kafka: %w", err)
		}

		w = kafka.NewWriter(kafka.WriterConfig{
			Brokers: option.Broker,
			Topic:   topic,
		})

		newConnection := connection{
			w:      w,
			broker: option.Broker,
			once:   sync.Once{},
		}

		newConnection.once.Do(func() {
			go func() {
				for {
					_, err := watcher.Next()
					if err != nil {
						time.Sleep(time.Second)
						continue
					}

					writers.Delete(topic)
					time.Sleep(10 * time.Second)
					newConnection.w.Close()
					return
				}
			}()
		})

		writers.Store(topic, newConnection)
		return w, nil
	}
	return c.(connection).w, nil
}

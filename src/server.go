package main

import (
	"os"
	"time"

	gitee_utils "gitee.com/sunmao-dx/strategy-executor/src/gitee-utils"
	"github.com/sirupsen/logrus"
)

func doRabbitMQ() {

	RMQ_QUEUE_NAME := os.Getenv("RMQ_QUEUE_NAME")
	RMQ_HOST := os.Getenv("RMQ_HOST")
	RMQ_VHOST := os.Getenv("RMQ_VHOST")
	RMQ_USER := os.Getenv("RMQ_USER")
	RMQ_PASS := os.Getenv("RMQ_PASS")
	RMQ_PORT := os.Getenv("RMQ_PORT")
	RMQ_ROUTINGKEY := os.Getenv("RMQ_ROUTINGKEY")
	RMQ_EXCHANGE_NAME := os.Getenv("RMQ_EXCHANGE_NAME")
	RMQ_EXCHANGE_TYPE := os.Getenv("RMQ_EXCHANGE_TYPE")
	RMQ_SCHEMA := "amqp"
	RMQ_CONNAME := ""

	//RabbitMQ
	rc := gitee_utils.RabbitConfig{
		Schema:         RMQ_SCHEMA,
		Username:       RMQ_USER,
		Password:       RMQ_PASS,
		Host:           RMQ_HOST,
		Port:           RMQ_PORT,
		VHost:          RMQ_VHOST,
		ConnectionName: RMQ_CONNAME,
	}
	rbt := gitee_utils.NewRabbit(rc)
	if err := rbt.Connect(); err != nil {
		gitee_utils.LogInstance.WithFields(logrus.Fields{
			"context": "Rabbitmq connect error",
		}).Info("info log")
		os.Exit(1)
	}
	cc := gitee_utils.ConsumerConfig{
		ExchangeName:  RMQ_EXCHANGE_NAME,
		ExchangeType:  RMQ_EXCHANGE_TYPE,
		RoutingKey:    RMQ_ROUTINGKEY,
		QueueName:     RMQ_QUEUE_NAME,
		ConsumerCount: 1,
		PrefetchCount: 1,
	}
	cc.Reconnect.MaxAttempt = 60
	cc.Reconnect.Interval = 1 * time.Second
	csm := gitee_utils.NewConsumer(cc, rbt)
	if err := csm.Start(); err != nil {
		gitee_utils.LogInstance.WithFields(logrus.Fields{
			"context": "Consumer setup error",
		}).Info("info log")
		os.Exit(1)
	}
	gitee_utils.LogInstance.WithFields(logrus.Fields{
		"context": "Consumer action success",
	}).Info("info log")
	select {}
}

func main() {
	gitee_utils.ConfigFile()
	doRabbitMQ()
}

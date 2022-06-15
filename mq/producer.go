package mq

import (
	"log"

	"github.com/junkeWu/filestore-server/config"
	util "github.com/junkeWu/filestore-server/utils"
	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

// 如果异常关闭，会接收通知
var notifyClose chan *amqp.Error

func init() {
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}
	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	// 判断是否创建过
	if channel != nil {
		return true
	}
	// 获得连接
	var err error
	conn, err = amqp.Dial(config.RabbitURL)
	util.Must(err)
	// 打开一个channel，用于消息的发布和接受
	channel, err = conn.Channel()
	util.Must(err)
	return true
}

// Publish 发布消息
func Publish(exchange, routingKey string, msg []byte) bool {
	// 判断channel是否正常
	if initChannel() == false {
		return false
	}
	// 执行消息发布动作
	err := channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)
	util.Must(err)
	return true
}

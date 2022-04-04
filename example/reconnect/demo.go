package main

import (
	"github.com/blackhades/go-amqp-lib/rabbitmq"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

func main() {
	conn, err := rabbitmq.Dial("amqp://127.0.0.1:5672")
	if err != nil {
		log.Panic(err)
	}

	sendCh, err := conn.Channel()
	if err != nil {
		log.Panic(err)
	}

	exchangeName := "test-exchange"
	queueName := "test-queue"

	err = sendCh.ExchangeDeclare(exchangeName, amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		log.Panic(err)
	}

	_, err = sendCh.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Panic(err)
	}

	if err := sendCh.QueueBind(queueName, "", exchangeName, false, nil); err != nil {
		log.Panic(err)
	}

	go func() {
		for {
			err := sendCh.Publish(exchangeName, "", false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(time.Now().String()),
			})
			log.Printf("publish, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()

	consumeCh, err := conn.Channel()
	if err != nil {
		log.Panic(err)
	}

	go func() {
		d, err := consumeCh.Consume(queueName, "", false, false, false, false, nil)
		if err != nil {
			log.Panic(err)
		}

		for msg := range d {
			log.Printf("msg: %s", string(msg.Body))
			msg.Ack(true)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	wg.Wait()
}

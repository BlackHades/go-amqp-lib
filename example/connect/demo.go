package main

import (
	"github.com/blackhades/go-amqp-lib/rabbitmq"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

func main() {
	url := "amqp://127.0.0.1:5672,amqp://127.0.0.1:5672,amqp://127.0.0.1:5672,amqp://127.0.0.1:5672"
	channel := rabbitmq.Connect(url)

	if channel == nil {
		log.Panic("Unable to connect to rabbitmq: ")
	}
	exchangeName := "test-exchange"
	queueName := "test-queue"

	queues := []string{
		"test-queue-1",
		"test-queue-2",
		"test-queue-3",
	}

	//Declare Multiple Queues with the same configuration
	err := rabbitmq.DeclareQueues(queues, true, false, false, false, nil)
	if err != nil {
		log.Panic(err)
	}

	exchanges := []string{
		"test-exchange-1",
		"test-exchange-2",
		"test-exchange-3",
	}

	err = rabbitmq.DeclareExchanges(exchanges, amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		log.Panic(err)
	}

	//Pushing to a queue using default mode
	go func() {
		for {
			err = channel.Publish("", queueName, false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("Pushing to a Queue Using Default - " + time.Now().String()),
			})
			log.Printf("publish, err: %v", err)
			time.Sleep(3 * time.Second)
		}
	}()

	//Pushing to a queue using the helper function
	go func() {
		for {
			err = rabbitmq.PublishToQueue(queueName, "Pushing to a Queue Using Helper - " + time.Now().String())
			log.Printf("publish, err: %v", err)
			time.Sleep(3 * time.Second)
		}
	}()

	//Pushing to an exchange using default mode
	go func() {
		for {
			err = channel.Publish(exchangeName, "", false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("Pushing to an Exchange using default - " + time.Now().String()),
			})
			log.Printf("publish, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()

	//Pushing to an exchange using helper function
	go func() {
		for {
			err = rabbitmq.PublishToExchange(exchangeName, "Pushing to an Exchange using helper - " + time.Now().String())
			log.Printf("publish, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()


	consumeCh, err := rabbitmq.GetConnection().Channel()
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

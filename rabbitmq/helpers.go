package rabbitmq

import (
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var channel *Channel
var connection *Connection
var once = new(sync.Once)
var DeadLetterQueue string

type GenericFN func([]byte) error

func resolveRabbitMQURL(url string) (error, []string) {
	if url == "" {
		url = os.Getenv("RABBITMQ_CLUSTER_URL")
	}

	if url == "" {
		url = os.Getenv("RABBITMQ_URL")
	}

	if url == "" {
		return errors.New("RabbitMQ URL is Required"), nil
	}
	return nil, strings.Split(url, ",")
}

func Connect(url string) *Channel {
	once.Do(func() {
		// Connect to the rabbitMQ instance
		err, urls := resolveRabbitMQURL(url)
		if err != nil {
			log.Panic(err)
		}

		var createdConnection *Connection
		if len(url) > 1 {
			log.Println("RabbitMQ: initiating cluster mode")
			createdConnection, err = DialCluster(urls)
		} else {
			createdConnection, err = Dial(urls[0])
		}

		if err != nil {
			log.Panic(err)
		}
		log.Println("Creating RabbitMQ Channel")
		// Create a channel from the connection. We'll use channels to access the data in the queue rather than the connection itself.
		createdChannel, err := createdConnection.Channel()
		if err != nil {
			log.Panic(err)
		}
		channel = createdChannel
		connection = createdConnection

		//init Queues and Exchanges

		err = channel.ExchangeDeclare("new-exchange", "fanout", true, false, false, false, nil)
		if err != nil {
			log.Println("Error Creating Exchange: DefaultServiceExchange", err)
		}
		//
		//fmt.Println("Error Creating exchange", err)
		//queue, err := channel.QueueDeclare(DefaultServiceQueue, true, false, false, false, nil)
		//log.Println("Queue", queue)
		//if err != nil {
		//	log.Println("Error Creating Queue: DefaultServiceQueue", err)
		//}
		//
		//queue, err = channel.QueueDeclare(DeadLetterQueue, true, false, false, false, nil)
		//log.Println("Queue", queue)
		//if err != nil {
		//	log.Println("Error Creating Queue: DeadLetterQueue", err)
		//}
		//
		//err = channel.QueueBind(DefaultServiceQueue, "", EventServiceExchange, false, nil)
		//if err != nil {
		//	log.Println("Error binding Queue: Event Service Exchange bind", err)
		//}
		//
		//err = channel.QueueBind(DefaultServiceQueue, "", DefaultServiceExchange, false, nil)
		//if err != nil {
		//	log.Println("Error binding Queue: Event Service Exchange bind", err)
		//}

		channel = createdChannel
		connection = createdConnection
		log.Println("Done Once")
	})
	return channel
}

func GetConnection() *Connection {
	return connection
}

func DeclareQueues(queues []string, durable bool, autoDelete bool, exclusive bool, noWait bool, table amqp.Table) error {
	for _, queue := range queues {
		_, err := channel.QueueDeclare(queue, durable, autoDelete, exclusive, noWait, table)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeclareExchanges(exchanges []string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, table amqp.Table) error {
	for _, exchange := range exchanges {
		err := channel.ExchangeDeclare(exchange, kind, durable, autoDelete, internal, noWait, table)
		if err != nil {
			return err
		}
	}
	return nil

}
func GetChannel() *Channel {
	return channel
}

func SetDeadLetterQueue(queueName string) {
	DeadLetterQueue = queueName
}

func PublishToExchange(exchangeName string, data interface{}) error {

	jsonPayload, _ := json.Marshal(data)
	//log.Println("QueuePayload", jsonPayload)
	message := amqp.Publishing{
		Body: []byte(string(jsonPayload)),
	}
	err := channel.Publish(exchangeName, "", false, false, message)
	if err != nil {
		log.Println("Error While Pushing to Exchange: ", err)
		return err
	}
	return nil
}

func PublishToQueue(queueName string, data interface{}) error {

	jsonPayload, _ := json.Marshal(data)
	message := amqp.Publishing{
		Timestamp: time.Now(),
		Body: []byte(string(jsonPayload)),
	}
	err := channel.Publish("", queueName, false, false, message)
	if err != nil {
		log.Println("Error While Pushing to Queue: ", err)
		return err
	}

	return nil
}

func Listen(listenMap map[string]GenericFN, prefetch int) {
	for queueName, functionToCall := range listenMap {
		go individualListen(queueName, functionToCall, prefetch)
	}
}

func individualListen(queueName string, functionToCall GenericFN, prefetch int) {

	channel, err := connection.Channel()
	if err != nil {
		panic("could not open channel with RabbitMQ:" + err.Error())
	}
	err = channel.Qos(
		prefetch, // prefetch count
		0,        // prefetch size
		false,    // global
	)

	log.Println("QOS Error", err)
	msgs, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		panic("Failed to register a consumer")
	}
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message on %s and pushing to goroutine", queueName)
			go processMessage(channel, d, functionToCall, queueName)
		}
	}()

	//>>>>>>> sandbox
	log.Println(" [*] Waiting for messages 0n " + queueName + ". To exit press CTRL+C")
	<-forever

	log.Println("Exiting...")
}

func processMessage(channel *Channel, delivery amqp.Delivery, functionToCall GenericFN, queueName string) {
	err := functionToCall(delivery.Body)
	if err != nil {
		log.Printf("Error Processing Message: %s\n", delivery.Body)
		//push to dead letter Queue,
		//Map result to slice
		jsonPayload, _ := json.Marshal(&map[string]interface{}{
			"reason":  err.Error(),
			"error":   err.Error(),
			"queue":   queueName,
			"payload": delivery.Body,
		})
		//log.Println("QueuePayload", jsonPayload)
		message := amqp.Publishing{
			Body: []byte(string(jsonPayload)),
		}
		_ = channel.Publish("", DeadLetterQueue, false, false, message)
	}
	_ = delivery.Ack(false)
	log.Println("Message Acknowledge")
}

func ProcessDeadLetter(payload []byte) error {
	deadPayload := map[string]interface{}{}
	_ = json.Unmarshal(payload, &deadPayload)
	log.Println("dead", deadPayload)

	return PublishToQueue(deadPayload["queue"].(string), deadPayload["payload"])
	//return nil
}

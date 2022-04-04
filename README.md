# streadway/amqp Conneciton/Channel auto reconnect wrap
streadway/amqp Connection/Channel does not reconnect if rabbitmq server restart/down and no cluster supported.

To simply developers, here is auto reconnect wrap with detail comments.

Also, with some basic helper functions like `Connect`, `DeclareQueues`, `DeclareExchanges`

This library depends on `streadway/amqp` and `sirius1024/go-amqp-reconnect`

## How to change existing code
1. add import `import "github.com/blackhades/go-amqp-reconnect/rabbitmq"`
2. Replace `amqp.Connection`/`amqp.Channel` with `rabbitmq.Connection`/`rabbitmq.Channel`

## Example
### Connect
> go run example/connect/demo.go

### Close by developer
> go run example/close/demo.go


### Auto reconnect
> go run example/reconnect/demo.go

### RabbitMQ Cluster with Reconnect
```go
import "github.com/blackhades/go-amqp-lib/rabbitmq"

rabbitmq.Connect("amqp://usr:pwd@127.0.0.1:5672/vhost,amqp://usr:pwd@127.0.0.1:5673/vhost,amqp://usr:pwd@127.0.0.1:5674/vhost"})
```

### To declare queues
```go
	url := "amqp://127.0.0.1:5672,amqp://127.0.0.1:5672,amqp://127.0.0.1:5672,amqp://127.0.0.1:5672"
	channel := rabbitmq.Connect(url)

	if channel == nil {
		log.Panic("Unable to connect to rabbitmq")
	}
	
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

```
### Publish/Push to a Queue
```go
//this will push to the queue every 3 seconds
    ...
	go func() {
		for {
			err = rabbitmq.PublishToQueue(queueName, "Pushing to a Queue Using Helper - " + time.Now().String())
			log.Printf("publish, err: %v", err)
			time.Sleep(3 * time.Second)
		}
	}()

```
### To declare an exchange
```go
	url := "amqp://127.0.0.1:5672,amqp://127.0.0.1:5672,amqp://127.0.0.1:5672,amqp://127.0.0.1:5672"
	channel := rabbitmq.Connect(url)

	if channel == nil {
		log.Panic("Unable to connect to rabbitmq")
	}
	
	exchanges := []string{
		"test-exchange-1",
		"test-exchange-2",
		"test-exchange-3",
	}

	//Declare Multiple Exchanges with the same configuration
	err := rabbitmq.DeclareExchanges(exchanges, amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		log.Panic(err)
	}
```
### Publish to an exchange
```go
	go func() {
		for {
			err = rabbitmq.PublishToExchange(exchangeName, "Pushing to an Exchange using helper - " + time.Now().String())
			log.Printf("publish, err: %v", err)
			time.Sleep(5 * time.Second)
		}
	}()
```

### To read/consume from a queue
```go
//Declare the function to process the data

func ProcessQueueData(byte []byte) error {}


//Map the queue to the function
    go rabbitmq.Listen(map[string]rabbitmq.GenericFN{
    "test-queue": ProcessQueueData,
    }, 1)
```

### To Set a dead letter queue
- create the dead letter queue
- if an error is thrown when processing, the data goes to a deadletter queue.
```go
    rabbitmq.SetDeadLetterQueue("dead-letter-queue")
```

### To Process a dead letter queue
**NOTE: when this function is called, data in the dead letter queue are sent back to the main queue they came from to be reprocessed.**
```go

    rabbitmq.Listen(map[string]rabbitmq.GenericFN{
    "dead-letter-queue": rabbitmq.ProcessDeadLetter,
    }, prefetch)

```
### Set reconnect delay seconds

You can set reconnect delay seconds with environment variable "**AMQP_RECONNECTION_DELAY_IN_SECONDS**"
e.g. 
```sh
# Default to 3s.
export AMQP_RECONNECTION_DELAY_IN_SECONDS=10
```
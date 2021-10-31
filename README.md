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

### Set reconnect delay seconds

You can set reconnect delay seconds with environment variable "**AMQP_RECONNECTION_DELAY_IN_SECONDS**"
e.g. 
```sh
# Default to 3s.
export AMQP_RECONNECTION_DELAY_IN_SECONDS=10
```
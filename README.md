### Overview

This repo tries to solve the problem of rabbitmq connection recovery. Basically it wraps functionality of `publishing`
and `consuming` of original "github.com/rabbitmq/amqp091-go" with connection recovery logic.

I assume that you have a good understanding of rabbitmq and leave you a full example of how it may look like in your
project in `cmd/auto_recovery` directory. Otherwise, start learning here: https://www.rabbitmq.com/getstarted.html

### Install

```shell
go get -u github.com/komron-m/rabbitmq
```

### Contributing

Pull requests are welcome. For any changes, please open an issue first to discuss what you would like to change.

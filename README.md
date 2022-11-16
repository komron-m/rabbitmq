### Overview

This repo tries to solve the problem of rabbitmq connection recovery, no matter how many rabbitmq nodes(clusters)
you have. Basically it wraps functionality of `publishing` and `consuming` of original "github.com/rabbitmq/amqp091-go"
with connection recovery logic.

I assume that you have a good understanding of rabbitmq and leave you a full example of how it may look like in your
project in `cmd/auto_recovery` directory.
Otherwise, start learning here: https://www.rabbitmq.com/getstarted.html

### Install

```shell
go get -u github.com/komron-m/rabbitmq
```

### Getting started locally

```shell
# clone repo
mkdir -p $GOPATH/src/github.com/komron-m/
cd $GOPATH/src/github.com/komron-m/
git clone git@github.com:komron-m/rabbitmq.git
cd rabbitmq

# only owner should have an access to this file, otherwise, rabbitmq will not start
chmod 600 etc/.erlang.cookie

# start docker with multiple (5) rabbitmq nodes
docker-compose up -d

# join all rabbitmq instance as a cluster, see: https://www.rabbitmq.com/clustering.html
# or ignore clustering and experiment with any desired instance
# run example app
go run ./cmd/auto_recovery

```

### Contributing

Pull requests are welcome. For any changes, please open an issue first to discuss what you would like to change.

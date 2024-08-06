# RMQ

a cli for `Apache RocketMQ` to manage topics, groups, clusters, acls, brokers, etc.


## Features

- topic
    - [x] create
    - [x] list
    - [x] produce
    - [x] describe
    - [ ] delete
    - [ ] update
- group
    - [x] create
    - [x] consume

## How to use

* Create a rocketmq config file `~/.config/rmq/rmq.yaml`:

```yaml
AccessKey: rocketmq2
SecretKey: 12345678
NamesrvAddrs: 127.0.0.1:9876
```

* Create a topic:

```shell
rmq topic create -t TopicTest -b 127.0.0.1:10911
```

* Describe a topic:

```shell
rmq topic describe -t TopicTest
```

* List all topics:

```shell
rmq topic list
```

* Produce a message:

```shell
rmq topic produce -t TopicTest
```

* Create a group:

```shell
rmq group create -g GroupTest
```

* Consume a message:

```shell
rmq group consume -g GID_Test -t TopicTest
```
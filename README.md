# RMQ

a cli for `Apache RocketMQ` to manage topics, groups, clusters, acls, brokers, etc.


## Features

- topic
    - [x] create
    - [x] produce
    - [ ] delete
    - [ ] update
    - [ ] describe

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

* Produce a message:

```shell
rmq topic produce -t TopicTest
```
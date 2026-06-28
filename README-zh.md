# Rmq

![](https://github.com/echooymxq/rmq/actions/workflows/build.yml/badge.svg?branch=main)

语言：[English](README.md) | 中文

[Apache RocketMQ](https://rocketmq.apache.org/) 命令行工具。

`rmq` 是面向 RocketMQ 用户的一款现代命令行工具，它按 Topic、Consumer Group、Message、Broker、Cluster、NameServer 等资源纬度组织常用操作。

`rmq` 支持使用Context同时管理多个 RocketMQ 集群，避免每次重复输入 NameServer 和 ACL 参数。

## 为什么选择 rmq

| 对比项 | rocketmq admin tool | rmq |
| --- | --- | --- |
| 连接配置 | 通常需要反复传递 NameServer 和 ACL 参数，或者在工具外部维护。 | 将集群配置保存为 Context，并通过 `rmq context use NAME` 切换；命令行参数仍可用于单次覆盖。 |
| 命令模型 | 命令名更接近底层 admin 操作。 | 按 `topic`、`group`、`message`、`broker`、`cluster`、`nameserver` 等资源组织命令。 |
| 输出体验 | 更偏向原始 admin 数据和 Java 侧诊断信息。 | 输出安静、稳定，表格和字段更便于阅读。 |
| Consumer 排障 | 连接、offset、lag、运行状态、线程栈通常需要通过多个 admin 命令分别查看。 | `rmq group connections`、`rmq group status`、`rmq group lag` 聚合展示日常诊断所需的关键信息。 |
| Message 排障 | 查询消息和确认消费组消费状态通常是两个独立流程。 | `rmq message query -t TOPIC -m MESSAGE_ID -g GROUP` 可以同时展示消息和 `ConsumeStatus`。 |
| 运维默认值 | 更直接暴露底层参数。 | 默认行为和命令说明更明确，例如 `topic produce` 未指定 body 时默认发送 4KB 随机报文，`group delete` 默认删除消费组 offset。 |

## Context 配置

Context 表示一个 RocketMQ 集群的连接配置。它包含集群的 NameServer 地址，以及可选的 ACL 凭据。你可以为 `local`、`staging`、`prod` 等环境维护多个 Context，并在命令之间切换，避免每次都重复传递连接参数。

命令行参数仍可用于单次覆盖，并且优先级高于当前选中的 Context。

通用参数：

| 参数 | 说明 |
| --- | --- |
| `--config` | 配置文件路径，默认为 `~/.config/rmq.yaml`。 |
| `--context` | 为当前命令选择指定 Context。 |
| `-n, --nameserver` | 逗号分隔的 NameServer 地址。 |
| `-a, --accessKey` | RocketMQ ACL access key。 |
| `-s, --secretKey` | RocketMQ ACL secret key。 |

Context 配置示例：

```yaml
current: local

contexts:
  local:
    namesrvAddrs:
      - localhost:9876
  prod:
    namesrvAddrs:
      - 10.0.0.1:9876
    accessKey: xxx
    secretKey: xxx
```

| 命令 | 说明 |
| --- | --- |
| `rmq context list` | 列出已配置的 Context，并标记当前 Context。 |
| `rmq context current` | 打印当前 Context 名称。 |
| `rmq context use NAME` | 设置当前 Context。 |
| `rmq context add NAME -n NAMESERVER [-a ACCESS_KEY] [-s SECRET_KEY]` | 新增 Context。如果未提供凭据，`rmq` 会要求确认，因为这表示目标集群必须关闭 ACL。 |

示例：

```shell
rmq context add local -n localhost:9876
rmq context add prod -n 10.0.0.1:9876 -a xxx -s xxx
rmq context use prod
rmq --context local topic list
```

## 支持的命令

### Topic

管理 RocketMQ Topic 并发送测试消息。

| 命令 | 说明 |
| --- | --- |
| `rmq topic list` | 列出所有 Topic。 |
| `rmq topic create -t TOPIC [-b BROKER] [-m MESSAGE_TYPE] [-q QUEUE_NUM]` | 在所有 master broker 上创建 Topic；指定 `-b` 时只在单个 broker 上创建。`-q` 同时设置读写队列数。 |
| `rmq topic delete -t TOPIC [-b BROKER]` | 从所有 master broker 删除 Topic；指定 `-b` 时只从单个 broker 删除。 |
| `rmq topic describe -t TOPIC` | 展示 Topic 配置。 |
| `rmq topic describe -t TOPIC --route` | 展示 Topic 路由数据。 |
| `rmq topic describe -t TOPIC --stats` | 展示 Topic offset 统计。 |
| `rmq topic produce -t TOPIC [-b BODY] [-c COUNT]` | 向 Topic 发送一条或多条消息。 |

示例：

```shell
rmq topic list
rmq topic create -t OrderTopic -q 8
rmq topic describe -t OrderTopic --route
rmq topic produce -t OrderTopic -b '{"id":1}' -c 3
```

### Consumer Group

管理消费组并消费消息。

| 命令 | 说明 |
| --- | --- |
| `rmq group list` | 列出 broker 上的订阅组。 |
| `rmq group create -g GROUP` | 创建订阅组。 |
| `rmq group delete -g GROUP` | 从所有 master broker 删除订阅组，并删除该组的消费 offset。 |
| `rmq group describe -g GROUP` | 展示订阅组配置。 |
| `rmq group connections -g GROUP` | 展示消费组概要、活跃 Consumer 实例和订阅信息。也可以使用别名 `clients` 或 `instances`。 |
| `rmq group status -g GROUP [-c CLIENT_ID] [-s]` | 展示 Consumer 运行状态、分配队列状态和 RT/TPS。`-s` 会包含客户端线程栈。 |
| `rmq group lag -g GROUP [-t TOPIC]` | 按队列展示消费堆积。`LastTimestamp` 会以本地可读时间展示。 |
| `rmq group consume -g GROUP -t TOPIC [--verbose]` | 启动 push consumer，并以 JSON lines 打印消费到的消息。 |

`group status` 会先打印 `Consumers` 汇总，再打印共享订阅信息，最后按 Consumer 实例打印详细块。每个实例块会在一个 `Consumer Queues` 表格中合并展示队列 offset 和缓存详情，随后展示 RT/TPS；时间戳会以本地可读时间展示。

示例：

```shell
rmq group list
rmq group create -g GID_ORDER
rmq group delete -g GID_ORDER
rmq group status -g GID_ORDER
rmq group status -g GID_ORDER -c 10.0.0.2@12345 -s
rmq group connections -g GID_ORDER
rmq group lag -g GID_ORDER
rmq group lag -g GID_ORDER -t OrderTopic
rmq group consume -g GID_ORDER -t OrderTopic --verbose
```

### Message

查看消息。

| 命令 | 说明 |
| --- | --- |
| `rmq message query -t TOPIC -m MESSAGE_ID [-g GROUP]` | 通过 MessageId 查询消息。指定 `-g` 时，会额外展示根据 offset 推断出的 `ConsumeStatus`，取值为 `consumed`、`not_consumed` 或 `unknown`。 |

示例：

```shell
rmq message query -t TopicTest -m 0A00000100002A9F0000000000012345
rmq message query -t TopicTest -m 0A00000100002A9F0000000000012345 -g GID_ORDER
```

### Broker

查看 Broker 和 Broker 配置。

| 命令 | 说明 |
| --- | --- |
| `rmq broker list` | 列出 cluster、broker、broker ID 和 broker 地址。 |
| `rmq broker config -b BROKER` | 打印 broker 配置。 |
| `rmq broker config -b BROKER -k KEY -v VALUE` | 更新单个 broker 配置项。 |

示例：

```shell
rmq broker list
rmq broker config -b 10.0.0.1:10911
rmq broker config -b 10.0.0.1:10911 -k autoCreateTopicEnable -v false
```

### Cluster

查看集群拓扑。

| 命令 | 说明 |
| --- | --- |
| `rmq cluster list` | 列出集群 broker group、broker ID、角色和地址。 |

示例：

```shell
rmq cluster list
```

### NameServer

查看和更新 NameServer 配置。

| 命令 | 说明 |
| --- | --- |
| `rmq nameserver config -a NAMESERVER` | 打印 NameServer 配置。 |
| `rmq nameserver config -a NAMESERVER -k KEY -v VALUE` | 更新单个 NameServer 配置项。 |

示例：

```shell
rmq nameserver config -a 10.0.0.1:9876
rmq nameserver config -a 10.0.0.1:9876 -k configStorePath -v /path/to/config
```

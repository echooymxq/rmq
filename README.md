# Rmq

![](https://github.com/echooymxq/rmq/actions/workflows/build.yml/badge.svg?branch=main)

The CLI (Command Line Interface) for [Apache RocketMQ](https://rocketmq.apache.org/).

`rmq` is a command interaction tool for RocketMQ users and OPS teams. It provides a modern, resource-oriented command model for daily operations, multi-environment workflows, troubleshooting, and automation.

Compared with the official admin tools, `rmq` focuses on quiet and predictable output, first-class context-based connection configuration, and script-friendly structured data rather than exposing low-level admin APIs directly.

## Context

A context represents the connection configuration for one RocketMQ cluster. It
contains the cluster's NameServer addresses and optional ACL credentials. You can
keep multiple contexts for environments such as `local`, `staging`, and `prod`,
then switch between them without repeating connection flags on every command.

Command-line flags can still be used for one-off overrides. They take precedence
over the selected context.

Common flags:

| Flag | Description |
| --- | --- |
| `--config` | Config file path. Defaults to `~/.config/rmq.yaml`. |
| `--context` | Select a named context for one command. |
| `-n, --nameserver` | Comma-separated NameServer addresses. |
| `-a, --accessKey` | RocketMQ ACL access key. |
| `-s, --secretKey` | RocketMQ ACL secret key. |

Example context file:

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

| Command | Description |
| --- | --- |
| `rmq context list` | List configured contexts and mark the current one. |
| `rmq context current` | Print the current context name. |
| `rmq context use NAME` | Set the current context. |
| `rmq context add NAME -n NAMESERVER [-a ACCESS_KEY] [-s SECRET_KEY]` | Add a context. If credentials are omitted, `rmq` asks for confirmation because the cluster must have ACL disabled. |

Examples:

```shell
rmq context add local -n localhost:9876
rmq context add prod -n 10.0.0.1:9876 -a xxx -s xxx
rmq context use prod
rmq --context local topic list
```

## Supported Commands

### Topic

Manage RocketMQ topics and produce test messages.

| Command | Description |
| --- | --- |
| `rmq topic list` | List all topics. |
| `rmq topic create -t TOPIC [-b BROKER] [-m MESSAGE_TYPE]` | Create a topic on all master brokers, or on one broker when `-b` is set. |
| `rmq topic delete -t TOPIC [-b BROKER]` | Delete a topic from all master brokers, or from one broker when `-b` is set. |
| `rmq topic describe -t TOPIC` | Show topic configuration. |
| `rmq topic describe -t TOPIC --route` | Show topic route data. |
| `rmq topic describe -t TOPIC --stats` | Show topic offset statistics. |
| `rmq topic produce -t TOPIC [-b BODY] [-c COUNT]` | Send one or more messages to a topic. |

Examples:

```shell
rmq topic list
rmq topic create -t OrderTopic
rmq topic describe -t OrderTopic --route
rmq topic produce -t OrderTopic -b '{"id":1}' -c 3
```

### Consumer Group

Manage consumer groups and consume messages.

| Command | Description |
| --- | --- |
| `rmq group list` | List subscription groups found on brokers. |
| `rmq group create -g GROUP` | Create a subscription group. |
| `rmq group describe -g GROUP` | Show subscription group configuration. |
| `rmq group describe -g GROUP --showConnection` | Show active consumer client connections. |
| `rmq group consume -g GROUP -t TOPIC [--verbose]` | Start a push consumer and print consumed messages as JSON lines. |

Examples:

```shell
rmq group list
rmq group create -g GID_ORDER
rmq group describe -g GID_ORDER --showConnection
rmq group consume -g GID_ORDER -t OrderTopic --verbose
```

### Message

Inspect messages.

| Command | Description |
| --- | --- |
| `rmq message query -m MSG_ID` | Query a message by offset message ID. |

Example:

```shell
rmq message query -m 0A00000100002A9F0000000000012345
```

### Broker

Inspect brokers and broker configuration.

| Command | Description |
| --- | --- |
| `rmq broker list` | List clusters, brokers, broker IDs, and broker addresses. |
| `rmq broker config -b BROKER` | Print broker config. |
| `rmq broker config -b BROKER -k KEY -v VALUE` | Update one broker config value. |

Examples:

```shell
rmq broker list
rmq broker config -b 10.0.0.1:10911
rmq broker config -b 10.0.0.1:10911 -k autoCreateTopicEnable -v false
```

### NameServer

Inspect and update NameServer configuration.

| Command | Description |
| --- | --- |
| `rmq nameserver config -a NAMESERVER` | Print NameServer config. |
| `rmq nameserver config -a NAMESERVER -k KEY -v VALUE` | Update one NameServer config value. |

Examples:

```shell
rmq nameserver config -a 10.0.0.1:9876
rmq nameserver config -a 10.0.0.1:9876 -k configStorePath -v /path/to/config
```

# rmq

![](https://github.com/echooymxq/rmq/actions/workflows/build.yml/badge.svg?branch=main)

Languages: English | [中文](README-zh.md)

The CLI (Command Line Interface) for [Apache RocketMQ](https://rocketmq.apache.org/).

`rmq` is a modern command-line tool for RocketMQ users. It organizes common operations by resource dimensions such as Topic, Consumer Group, Message, Broker, Cluster, and NameServer.

`rmq` supports using Contexts to manage multiple RocketMQ clusters at the same time, avoiding repeated NameServer and ACL parameters on every command.

## Why rmq

| Area | rocketmq admin tool                                                                                            | rmq                                                                                                                                                                           |
| --- |----------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Connection config | Pass NameServer and ACL flags repeatedly, or manage them outside the tool.                                     | Store cluster configs as contexts and switch with `rmq context use NAME`. Command flags still work as one-off overrides.                                                      |
| Command model | Command names mostly mirror low-level admin operations.                                                        | Resource-oriented commands grouped by `topic`, `group`, `message`, `broker`, `cluster`, and `nameserver`.                                                                     |
| Output | Often optimized for raw admin data and Java-side diagnostics.                                                  | Quiet, predictable CLI output with readable tables, stable fields.                                                                                                            |
| Consumer troubleshooting | Connection, offset, lag, running status, and stack data are usually inspected through separate admin commands. | `rmq group connections`, `rmq group status`, and `rmq group lag` aggregate the high-signal consumer data needed for daily diagnosis.                                          |
| Message troubleshooting | Querying a message and checking its group consumption state are separate workflows.                            | `rmq message query -t TOPIC -m MESSAGE_ID -g GROUP` shows the message and its `ConsumeStatus` together.                                                                       |
| Operational defaults | Exposes more low-level switches directly.                                                                      | Uses clearer defaults and command descriptions, such as a 4KB random body for `topic produce` when no body is provided, and `group delete` removing group offsets by default. |

## Install

Install the latest release on macOS or Linux:

```shell
curl -fsSL https://raw.githubusercontent.com/echooymxq/rmq/main/scripts/install.sh | bash
```

The install script downloads the matching release package for your OS and architecture, then installs `rmq` to `/usr/local/bin` by default.

Install a specific version:

```shell
curl -fsSL https://raw.githubusercontent.com/echooymxq/rmq/main/scripts/install.sh | VERSION=0.1.0 bash
```

Install to a custom directory:

```shell
curl -fsSL https://raw.githubusercontent.com/echooymxq/rmq/main/scripts/install.sh | BINDIR="$HOME/.local/bin" bash
```

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
| `rmq topic create -t TOPIC [-b BROKER] [-m MESSAGE_TYPE] [-q QUEUE_NUM]` | Create a topic on all master brokers, or on one broker when `-b` is set. `-q` sets both read and write queue counts. |
| `rmq topic delete -t TOPIC [-b BROKER]` | Delete a topic from all master brokers, or from one broker when `-b` is set. |
| `rmq topic describe -t TOPIC` | Show topic configuration. |
| `rmq topic describe -t TOPIC --route` | Show topic route data. |
| `rmq topic describe -t TOPIC --stats` | Show topic offset statistics. |
| `rmq topic produce -t TOPIC [-b BODY] [-c COUNT]` | Send one or more messages to a topic. |

Examples:

```shell
rmq topic list
rmq topic create -t OrderTopic -q 8
rmq topic describe -t OrderTopic --route
rmq topic produce -t OrderTopic -b '{"id":1}' -c 3
```

### Consumer Group

Manage consumer groups and consume messages.

| Command | Description |
| --- | --- |
| `rmq group list` | List subscription groups found on brokers. |
| `rmq group create -g GROUP` | Create a subscription group. |
| `rmq group delete -g GROUP` | Delete a subscription group from all master brokers and remove the group's consumer offsets. |
| `rmq group describe -g GROUP` | Show subscription group configuration. |
| `rmq group connections -g GROUP` | Show group summary, active consumer instances and subscriptions. Also available as `clients` or `instances`. |
| `rmq group status -g GROUP [-c CLIENT_ID] [-s]` | Show consumer running status, assigned queue state and RT/TPS. `-s` includes the client stack dump. |
| `rmq group lag -g GROUP [-t TOPIC]` | Show consumer lag by queue. `LastTimestamp` is shown in local readable time. |
| `rmq group consume -g GROUP -t TOPIC [--verbose]` | Start a push consumer and print consumed messages as JSON lines. |

`group status` prints a `Consumers` summary first, then shared subscriptions, then one detailed block per consumer instance. Per-instance blocks include queue offsets and cache details in one `Consumer Queues` table, followed by RT/TPS; timestamps are rendered in local readable time.

Examples:

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

Inspect messages.

| Command | Description |
| --- | --- |
| `rmq message query -t TOPIC -m MESSAGE_ID [-g GROUP]` | Query a message by MessageId. With `-g`, also shows `ConsumeStatus`, inferred from offsets as `consumed`, `not_consumed`, or `unknown`. |

Example:

```shell
rmq message query -t TopicTest -m 0A00000100002A9F0000000000012345
rmq message query -t TopicTest -m 0A00000100002A9F0000000000012345 -g GID_ORDER
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

### Cluster

Inspect cluster topology.

| Command | Description |
| --- | --- |
| `rmq cluster list` | List cluster broker groups, broker IDs, roles, and addresses. |

Examples:

```shell
rmq cluster list
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

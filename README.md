# Rmq

![](https://github.com/echooymxq/rmq/actions/workflows/build.yml/badge.svg?branch=main)

The CLI (Command Line Interface) for [Apache RocketMQ](https://rocketmq.apache.org/).

`rmq` is a command interaction tool for RocketMQ users and OPS teams. It provides a modern, resource-oriented command model for daily operations, multi-environment workflows, troubleshooting, and automation.

Compared with the official admin tools, `rmq` focuses on quiet and predictable output, first-class context-based connection configuration, and script-friendly structured data rather than exposing low-level admin APIs directly.

## Features

- topic
    - [x] create
    - [x] list
    - [x] produce
    - [x] describe
    - [ ] delete
    - [ ] update
    - [ ] route
- group
    - [x] create
    - [x] list
    - [x] describe
    - [x] consume
    - [ ] delete
    - [ ] update
- message
    - [x] query
    - [ ] trace
- nameserver
  - [x] config
- broker
  - [x] list
  - [x] config

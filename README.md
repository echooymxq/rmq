# RMQ

`rmq` aims to provide a modern RocketMQ CLI with a resource-oriented command model, first-class context-based connection configuration, quiet and predictable output, and script-friendly structured data. Compared with the official admin tools, it focuses on daily operations, multi-environment workflows, troubleshooting, and automation rather than exposing low-level admin APIs directly.


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

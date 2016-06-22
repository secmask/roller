### Introduction

_Roller_ implement simple pub-sub service like Redis PubSub (and also use redis protocol)
It's very simple and not provide all redis pub-sub feature, it designed to publish real time data.

- Only the first channel in `subscribe` is work, mean multiple channel subscribe in single connection is not support
- Each client has a temporary queue(default max length = 50k) hold its on delivery messages, if client cannot catch up, it will be drop(roller is designed for real time)

### License
Roller is provide under MIT License
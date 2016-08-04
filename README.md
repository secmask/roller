### Introduction
[![Circle CI](https://circleci.com/gh/secmask/roller.svg?style=svg&circle-token=fdfcf508f51b281788a87d8ed8288372d3b5f488)](https://circleci.com/gh/secmask/roller)

_Roller_ implement simple pub-sub service like Redis PubSub (and also use redis protocol)
It's very simple and not provide all redis pub-sub feature, it designed to publish real time data.

- Only the first channel in `subscribe` is work, mean multiple channel subscribe in single connection is not support
- Each client has a temporary queue(default max length = 50k) hold its on delivery messages, if client cannot catch up, it will be drop(roller is designed for real time)

### License
Roller is provide under MIT License
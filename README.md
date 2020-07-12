# go-ws-websocket

`go-ws-websocket` is a package that Websocket clients which are registered a broker broadcast messages.

<p align="left">       
   <a href="https://hits.seeyoufarm.com"/><img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fgjbae1212%2Fgo-ws-broadcast"/></a>
   <a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-GREEN.svg" alt="license"/></a>
</p>

## Installation
```bash
$ go get -u github.com/gjbae1212/go-ws-websocket
```

## Usage
```go
# create braker
broker, _ :=  NewBreaker()

# register 
client, _ := broker.Register(websocket)

# unregister 
_ = broker.Register(client)

# broadcast message
_ = broker.BroadCast(message)   
```

## LICENSE
This project is following The MIT.

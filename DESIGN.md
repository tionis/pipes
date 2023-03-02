# Architecture Design
## Client Perspective
- Clients connect to an interactive socket (either a ssh-key authenticated websocket or a ssh connection).
- They then can send and receive either msgpack or json-encoded messages
- They can send messages to listen to specific channels, the exact behaviour of them depends on their type
- Fifo channels ensure
- Request channles are special fifo channels with a `.req` ending like for example `example/channel.req`. A consumer can consume them and then write their answer into the `example/channel.res` to respond to the request
- The built-in webhook support writes a json file containing all data about the incoming http request to a request channel and returns the answer it receives on the response channel back to the client.
- Stream channels can be openend with or without a timestamp. If a timestamp is given the client will receive all stored and future messages after (not including) that timestamp, if none is given the current time is used as timestamp which means the client will simply receive all future messages.
- Pubsub channels work in a fanout pattern

# gipc

[![Testing](https://github.com/joe-at-startupmedia/gipc/actions/workflows/testing.yml/badge.svg)](https://github.com/joe-at-startupmedia/gipc/actions/workflows/testing.yml)
[![codecov](https://codecov.io/gh/joe-at-startupmedia/gipc/graph/badge.svg?token=0G9FP0QN5S)](https://codecov.io/gh/joe-at-startupmedia/gipc)
[![Go Report Card](https://goreportcard.com/badge/github.com/joe-at-startupmedia/gipc)](https://goreportcard.com/report/github.com/joe-at-startupmedia/gipc)

### Overview
 
A simple-to-use package that creates a communication channel between two go processes.


## Usage

Create a server with the default configuration and start listening for the client:

```go
s, err := gipc.StartServer(&ServerConfig{Name:"<name of connection>"})
if err != nil {
	log.Println(err)
	return
}
```
Create a client and connect to the server:

```go
c, err := gipc.StartClient(&ClientConfig{Name:"<name of connection>"})
if err != nil {
	log.Println(err)
	return
}
```

### Read messages 

Read each message sent (blocking):

```go
for {

	//message, err := s.Read() // server
	message, err := c.Read() // client
	
	if err == nil {
	// handle error
	}
	
	// do something with the received messages
}
```

Read each message sent until a specific duration has surpassed. 

```go
for {

	message, err := c.ReadTimed(5*time.Second)
	
	if  message == gipc.TimeoutMessage {
		continue
    }   
	
	if err == nil && c.StatusCode() != gipc.Connecting {
	
	} 
}
```

### MultiClient Mode

Allow polling of newly created clients on each iteration until a specific duration has surpassed. 

```go
s, err := gipc.StartServer(&ServerConfig{Name:"<name of connection>", MultiClient: true})
    if err != nil {
    log.Println(err)
    return
}

for {
    s.Connections.ReadTimed(5*time.Second, func(srv *ipc.Server, message *ipc.Message, err error) {
        if  message == gipc.TimeoutMessage {
            continue
        }
        
        if message.MsgType == -1 && message.Status == "Connected" {
        
        }
    })
}
```

* `Server.Connections.ReadTimed` will block until the slowest ReadTimed callback completes. 
* `Server.Connections.ReadTimedFastest` will unblock after the first ReadTimed callback completes.

While `ReadTimedFastest` will result in faster iterations, it will also result in more running goroutines in scenarios where clients requests are not evenly distributed. 

To get a better idea of how these work, run the following examples: 

Using `ReadTimed`:
```bash
go run --race example/multiclient/multiclient.go
```

Using `ReadTimedFastest`: 
```bash
FAST=true go run --race example/multiclient/multiclient.go
```

Notice that the Server receives messages faster and the process will finish faster

### Message Struct

All received messages are formatted into the type Message

```go
type Message struct {
	Err     error  // details of any error
	MsgType int    // 0 = reserved , -1 is an internal message (disconnection or error etc), all messages recieved will be > 0
	Data    []byte // message data received
	Status  string // the status of the connection
}
```

### Write a message


```go

//err := s.Write(1, []byte("<Message for client"))
err := c.Write(1, []byte("<Message for server"))

if err == nil {
// handle error
}
```

 ## Advanced Configuration

Server options:

```go
config := &gipc.ServerConfig{
	Name: (string),            // the name of the queue (required)
	Encryption: (bool),        // allows encryption to be switched off (bool - default is true)
	MaxMsgSize: (int) ,        // the maximum size in bytes of each message ( default is 3145728 / 3Mb)
	UnmaskPermissions: (bool), // make the socket writeable for other users (default is false)
	MultiMode: (bool),         // allow the server to connect with multiple clients
}
```

Client options:

```go
config := &gipc.ClientConfig  {
	Name: (string),             // the name of the queue needs to match the name of the ServerConfig (required)
	Encryption: (bool),         // allows encryption to be switched off (bool - default is true)
	Timeout: (time.Duration),   // duration to wait while attempting to connect to the server (default is 0 no timeout)
	RetryTimer: (time.Duration),// duration to wait before iterating the dial loop or reconnecting (default is 1 second)
}
```

By default, the `Timeout` value is 0 which allows the dial loop to iterate in perpetuity until a connection to the server is established. 

In scenarios where a perpetually attempting to reconnect is impractical, a `Timeout` value should be provided. When the connection times out, no further retries will be attempted. 

When a Client is no longer used, ensure that the `.Close()` method is called to prevent unnecessary perpetual connection attempts.

 ### Encryption

 By default, the connection established will be encrypted, ECDH384 is used for the key exchange and AES 256 GCM is used for the cipher.

 Encryption can be switched off by passing in a custom configuration to the server & client start function:

```go
Encryption: false
```

 ### Unix Socket Permissions

Under most configurations, a socket created by a user will by default not be writable by another user, making it impossible for the client and server to communicate if being run by separate users. The permission mask can be dropped during socket creation by passing a custom configuration to the server start function.  **This will make the socket writable for any user.**

```go
UnmaskPermissions: true	
```

## TCP Support

Instead of using Unix domain sockets, you can also use TCP. This provides the benefits from TCP reliability and platform interoperability (i.e. Windows) but also sacrifices performance and cpu/memory.

To build with TCP support:
```bash
go build -tags network
```

You can customize the following using runtime environment variables:
* `GIPC_NETWORK_HOST`: The address host of which the TCP connection is bound to, by default this is 127.0.0.1
* `GIPC_NETWORK_PORT`: The address port of which the TCP connection is bound to, by default this is 8100

```bash
GIPC_NETWORK_HOST=10.0.2.15 GIPC_NETWORK_PORT=7200 go run -tags network
```

## Debugging

### Environment Variables

You can specify debug verbosity using the `GIPC_DEBUG` environment variable.

```bash
GIPC_DEBUG=true make run
```

`GIPC_DEBUG` accepts the following values:
* `true`: sets the debug level to debug
* `debug`: has the same effect as true
* `info`: sets the debug level to info
* `warn`: sets the debug level to warn
* `error`: sets the debug level to error

## Testing

The package has been tested on Mac and Linux and has extensive test coverage. The following commands will run all the tests and examples with race condition detection enabled.

```bash
make test run
```

You can change the speed of the tests by providing a value for the `GIPC_WAIT` environment variable. A value `> 5` will specify the amount of milliseconds to wait in between critical intervals whereas a value `<= 5` will resolve to the amount of seconds to wait in between the same. The default value is 10 milliseconds. You can also provide the `IPC_DEBUG=true` environment variable to set the `logrus.Loglevel` to debug mode. The following command will make the tests run in debug mode while waiting 500ms in between critical intervals:

```bash
GIPC_WAIT=500 GIPC_DEBUG=true make test run
```

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joe-at-startupmedia/gipc"
)

const CONN_NAME = "example_multi"

func main() {

	s := server()
	defer s.Close()

	// change the sleep time by using GIPC_WAIT env variable (seconds)
	gipc.Sleep()

	clientConfig := &gipc.ClientConfig{Name: CONN_NAME, MultiClient: true, Encryption: gipc.ENCRYPT_BY_DEFAULT}
	c1, err := gipc.StartClient(clientConfig)
	if err != nil {
		panic(err)
	}
	defer c1.Close()

	gipc.Sleep()

	c2, err := gipc.StartClient(clientConfig)
	if err != nil {
		panic(err)
	}
	defer c2.Close()

	gipc.Sleep()

	c3, err := gipc.StartClient(clientConfig)
	if err != nil {
		panic(err)
	}
	defer c3.Close()

	serverPonger(c2, false)
	//time.Sleep(6 * time.Second)

	gipc.Sleep()

	wg := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		serverPonger(c1, false)
		gipc.Sleep()
		wg <- true
	}()

	serverPonger(c3, false)

	gipc.Sleep()

	serverPonger(c2, true)

	gipc.Sleep()

	<-wg
}

func serverPonger(c *gipc.Client, autosend bool) {

	pongMessage := fmt.Sprintf("Message from client(%d) - PONG", c.ClientId)

	if autosend {
		c.Write(5, []byte(pongMessage))
		return
	}

	for {

		message, err := c.ReadTimed(5 * time.Second)

		if message == gipc.TimeoutMessage {
			continue
		} else if err != nil {
			log.Println("Client Read err: ", err)
			if err.Error() == "Client.Read timed out" {
				panic(err)
			}
			continue
		}

		if message.MsgType == -1 { //internal message

			log.Println("client status", c.Status())

			if message.Status == "Reconnecting" {
				panic("Reconnecting")
			}

		} else { //user message
			log.Printf("Client(%d) received: %s - Message type: %d", c.ClientId, string(message.Data), message.MsgType)
			err2 := c.Write(5, []byte(pongMessage))
			if err2 != nil {
				log.Println("Client Write  err: ", err)
			}
			break
		}

		gipc.Sleep()
	}

}

func server() *gipc.Server {

	s, err := gipc.StartServer(&gipc.ServerConfig{Name: CONN_NAME, MultiClient: true, Encryption: gipc.ENCRYPT_BY_DEFAULT})
	if err != nil {
		panic(err)
	}

	startTime := time.Now()
	useFastest := os.Getenv("FAST") == "true"

	go func() {
		for {

			// we need to use the ReadTimed in order to poll all new clients
			if useFastest {
				s.Connections.ReadTimedFastest(5*time.Second, onReadTimedFinish)
			} else {
				s.Connections.ReadTimed(5*time.Second, onReadTimedFinish)
			}

			fmt.Printf("server.ReadTimed next iteration after (%s) seconds since start \n", time.Since(startTime))
		}
	}()

	return s
}

func onReadTimedFinish(srv *gipc.Server, message *gipc.Message, err error) {
	if message == gipc.TimeoutMessage {
		return
	} else if err != nil {
		log.Println("Read err: ", err)
		return
	}

	if message.MsgType == -1 { //internal message

		log.Println("server status", srv.Status())

		if message.Status == "Connected" {
			log.Println("server sending ping: status", srv.Status())
			srv.Write(1, []byte("server - PING"))
		}

	} else { //user message

		log.Println("Server received: "+string(message.Data)+" - Message type: ", message.MsgType)
	}
}

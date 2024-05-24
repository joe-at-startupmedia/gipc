//go:build shmem

package gipc

import (
	"fmt"
	"github.com/joe-at-startupmedia/shmemipc"
)

func getSocketName(clientId int, name string) string {
	if clientId > 0 {
		return fmt.Sprintf("%s%s%d", SOCKET_NAME_BASE, name, clientId)
	} else {
		return fmt.Sprintf("%s%s", SOCKET_NAME_BASE, name)
	}
}

func (c *Client) connect() (ActorConn, error) {

	socketName := getSocketName(c.ClientId, c.config.ClientConfig.Name)
	requester := shmemipc.NewRequester(socketName)
	c.requester = requester
	return requester.GetConn(), requester.GetError()
}

func (s *Server) listen(clientId int) error {

	socketName := getSocketName(clientId, s.config.ServerConfig.Name)

	memsize := uint64(s.config.ServerConfig.MaxMsgSize)
	//memsize := uint64(1024)

	responder := shmemipc.NewResponder(socketName, memsize)

	s.listener = responder
	s.responder = responder

	err := responder.GetError()
	if err != nil {
		return err
	}

	ServerConn <- responder.GetConn()

	return nil
}

var ServerConn = make(chan *shmemipc.IpcResponderConn, 1)

func (s *Server) accept() (ActorConn, error) {
	conn := <-ServerConn
	return conn, nil
}

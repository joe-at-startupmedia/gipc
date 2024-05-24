//go:build shmem

package gipc

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// 1st message sent from the server
// byte 0 = protocol VERSION no.
func (sc *Server) handshake() error {

	err := sc.one()
	if err != nil {
		return err
	}

	if sc.shouldUseEncryption() {
		err = sc.startEncryption()
		if err != nil {
			return err
		}
	}

	err = sc.msgLength()
	if err != nil {
		return err
	}

	return nil
}

func (sc *Server) one() error {

	buff := make([]byte, 2)

	buff[0] = byte(VERSION)

	if sc.shouldUseEncryption() {
		buff[1] = byte(1)
	} else {
		buff[1] = byte(0)
	}

	err := sc.responder.Write(buff)
	if err != nil {
		return errors.New("unable to send handshake ")
	}

	recv, err := sc.responder.Read()
	if err != nil {
		return errors.New("failed to received handshake reply")
	}

	switch result := recv[0]; result {
	case 0:
		return nil
	case 1:
		return errors.New("client has a different VERSION number")
	case 2:
		return errors.New("client is enforcing encryption")
	case 3:
		return errors.New("server failed to get handshake reply")
	}

	return errors.New("other error - handshake failed")
}

func (sc *Server) msgLength() error {

	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(sc.config.ServerConfig.MaxMsgSize))

	var err error

	if sc.shouldUseEncryption() {
		buff, err = encrypt(*sc.cipher, buff)
		if err != nil {
			return err
		}
	}

	toSend := make([]byte, 4)
	binary.LittleEndian.PutUint32(toSend, uint32(sc.config.ServerConfig.MaxMsgSize))
	toSend = append(toSend, buff...)

	err = sc.responder.Write(toSend)
	if err != nil {
		return errors.New("unable to send max message length ")
	}

	_, err = sc.responder.Read()
	if err != nil {
		return errors.New("did not received message length reply")
	}

	return nil
}

// 1st message received by the client
func (cc *Client) handshake() error {

	err := cc.one()
	if err != nil {
		return err
	}

	if cc.shouldUseEncryption() {
		err = cc.startEncryption()
		if err != nil {
			return err
		}
	}

	err = cc.msgLength()
	if err != nil {
		return err
	}

	return nil
}

func (cc *Client) one() error {

	recv, err := cc.requester.Read()
	if err != nil {
		return errors.New("failed to received handshake message")
	}

	if recv[0] != VERSION {
		cc.handshakeSendReply(1)
		return errors.New("server has sent a different VERSION number")
	}

	if recv[1] != 1 && cc.shouldUseEncryption() {
		cc.handshakeSendReply(2)
		return errors.New("server tried to connect without encryption")
	}

	return cc.handshakeSendReply(0)
}

func (cc *Client) msgLength() error {

	buff, err := cc.requester.Read()
	if err != nil {
		return errors.New("failed to received max message length 1")
	}

	var msgLen uint32
	err = binary.Read(bytes.NewReader(buff), binary.LittleEndian, &msgLen) // message length
	if err != nil {
		return errors.New("failed to read binary")
	}

	buff, err = cc.requester.Read()
	if err != nil {
		return errors.New("failed to received max message length 2")
	}

	if cc.shouldUseEncryption() {
		buff, err = decrypt(*cc.cipher, buff)
		if err != nil {
			return errors.New("failed to received max message length 3")
		}
	}

	var maxMsgSize uint32
	err = binary.Read(bytes.NewReader(buff), binary.LittleEndian, &maxMsgSize) // message length
	if err != nil {
		return errors.New("failed to read binary")
	}

	cc.maxMsgSize = int(maxMsgSize)
	return cc.handshakeSendReply(0)
}

func (cc *Client) handshakeSendReply(result byte) error {

	buff := make([]byte, 1)
	buff[0] = result

	err := cc.requester.Write(buff)
	return err
}

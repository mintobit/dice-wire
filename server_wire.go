// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wire

import (
	"fmt"
	"net"
	"os"
	"time"
	"wire/internal"
)

type ServerWire struct {
	*internal.ProtobufTCPWire
}

func NewServerWire(maxMsgSize int, keepAlive int, clientFD int) (*ServerWire, error) {
	file := os.NewFile(uintptr(clientFD), "client-connection")
	if file == nil {
		return nil, fmt.Errorf("failed to create file from file descriptor")
	}

	conn, err := net.FileConn(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create net.Conn from file descriptor: %w", err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			return nil, fmt.Errorf("failed to set TCP_NODELAY: %w", err)
		}
		if err := tcpConn.SetKeepAlive(true); err != nil {
			return nil, fmt.Errorf("failed to set keepalive: %w", err)
		}
		if err := tcpConn.SetKeepAlivePeriod(time.Duration(keepAlive) * time.Second); err != nil {
			return nil, fmt.Errorf("failed to set keepalive period: %w", err)
		}
	}

	wire := &ServerWire{
		ProtobufTCPWire: internal.NewProtobufTCPWire(maxMsgSize, conn),
	}

	return wire, nil
}

func (sw *ServerWire) Send(resp *Response) error {
	return sw.ProtobufTCPWire.Send(resp)
}

func (sw *ServerWire) Receive() (*Command, error) {
	cmd := &Command{}

	err := sw.ProtobufTCPWire.Receive(cmd)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (sw *ServerWire) Close() {
	sw.ProtobufTCPWire.Close()
}

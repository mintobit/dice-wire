// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wire

import (
	"fmt"
	"net"
	"time"
	"wire/internal"
)

type ClientWire struct {
	*internal.ProtobufTCPWire
}

func NewClientWire(maxMsgSize int, host string, port int) (*ClientWire, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	wire := &ClientWire{
		ProtobufTCPWire: internal.NewProtobufTCPWire(maxMsgSize, conn),
	}

	return wire, nil
}

func (cw *ClientWire) Send(cmd *Command) error {
	return cw.ProtobufTCPWire.Send(cmd)
}

func (cw *ClientWire) Receive() (*Response, error) {
	resp := &Response{}
	err := cw.ProtobufTCPWire.Receive(resp)

	return resp, err
}

func (cw *ClientWire) Close() {
	cw.ProtobufTCPWire.Close()
}

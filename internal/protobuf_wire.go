// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package internal

import (
	"google.golang.org/protobuf/proto"
	"net"
)

type ProtobufTCPWire struct {
	tcpWire Wire
}

func NewProtobufTCPWire(maxMsgSize int, conn net.Conn) *ProtobufTCPWire {
	return &ProtobufTCPWire{
		tcpWire: NewTCPWire(maxMsgSize, conn),
	}
}

func (w *ProtobufTCPWire) Send(msg proto.Message) error {
	buffer, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return w.tcpWire.Send(buffer)
}

func (w *ProtobufTCPWire) Receive(dst proto.Message) error {
	buffer, err := w.tcpWire.Receive()
	if err != nil {
		return err
	}

	err = proto.Unmarshal(buffer, dst)
	if err != nil {
		return err
	}

	return nil
}

func (w *ProtobufTCPWire) Close() {
	w.tcpWire.Close()
}

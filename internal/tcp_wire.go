// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package internal

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

const prefixSize = 4 // bytes

type TCPWire struct {
	maxMsgSize int
	readMu     sync.Mutex
	writeMu    sync.Mutex
	conn       net.Conn
}

func NewTCPWire(maxMsgSize int, conn net.Conn) *TCPWire {
	return &TCPWire{
		maxMsgSize: maxMsgSize,
		conn:       conn,
	}
}

func (w *TCPWire) Send(msg []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	size := len(msg)
	buffer := make([]byte, prefixSize+size)

	w.writePrefix(size, buffer)
	copy(buffer[prefixSize:], msg)

	_, err := w.conn.Write(buffer)

	return err
}

func (w *TCPWire) Receive() ([]byte, error) {
	w.readMu.Lock()
	defer w.readMu.Unlock()

	size, err := w.readPrefix()
	if err != nil {
		return nil, err
	}

	if size <= 0 {
		return nil, fmt.Errorf("invalid message size: %d", size)
	}

	if size > uint32(w.maxMsgSize) {
		return nil, fmt.Errorf("message too large: %d bytes (max: %d)", size, w.maxMsgSize)
	}

	buffer := make([]byte, size)

	_, err = io.ReadFull(w.conn, buffer)

	if err != nil {
		return nil, fmt.Errorf("failed to read message into buffer: %w", err)
	}

	return buffer, nil
}

func (w *TCPWire) Close() {
	err := w.conn.Close()
	if err != nil {
		// log error

		return
	}
}

func (w *TCPWire) readPrefix() (uint32, error) {
	buffer := make([]byte, prefixSize)

	_, err := io.ReadFull(w.conn, buffer)

	if err != nil {
		return 0, fmt.Errorf("failed to read prefix: %w", err)
	}

	return binary.BigEndian.Uint32(buffer), nil
}

func (w *TCPWire) writePrefix(msgSize int, buffer []byte) {
	binary.BigEndian.PutUint32(buffer[:prefixSize], uint32(msgSize))
}

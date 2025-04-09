// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package internal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
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

	_, err := write(w.conn, buffer)

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
		return nil, WireError{
			Kind: corruptMessage,
			Err:  fmt.Errorf("invalid message size: %d", size),
		}
	}

	if size > uint32(w.maxMsgSize) {
		return nil, WireError{
			Kind: corruptMessage,
			Err:  fmt.Errorf("message too large: %d bytes (max: %d)", size, w.maxMsgSize),
		}
	}

	buffer := make([]byte, size)

	_, err = read(w.conn, buffer)
	if err != nil {
		return nil, err
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

	_, err := read(w.conn, buffer)

	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(buffer), nil
}

func (w *TCPWire) writePrefix(msgSize int, buffer []byte) {
	binary.BigEndian.PutUint32(buffer[:prefixSize], uint32(msgSize))
}

func read(r io.Reader, buf []byte) (int, error) {
	n, err := io.ReadFull(r, buf)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, WireError{Kind: closedGracefully, Err: err}
		}
		var netErr net.Error
		if errors.As(err, &netErr) {
			if netErr.Timeout() {
				return 0, WireError{Kind: timeout, Err: netErr}
			}

			return 0, WireError{Kind: unknown, Err: err}
		}

		if strings.Contains(err.Error(), "use of closed network connection") {
			return 0, WireError{Kind: closedAbruptly, Err: err}
		}

		return 0, WireError{Kind: unknown, Err: err}
	}

	return n, nil
}

func write(conn net.Conn, buffer []byte) (int, error) {
	var totalWritten int
	const maxWriteAttempts = 3
	writeAttempts := 0

	for totalWritten < len(buffer) {
		if writeAttempts >= maxWriteAttempts {
			return totalWritten, WireError{
				Kind: partialWrite,
				Err:  fmt.Errorf("maximum retry limit reached, only %d bytes written", totalWritten),
			}
		}

		n, err := conn.Write(buffer[totalWritten:])
		if err != nil {
			var netErr net.Error
			switch {
			case errors.Is(err, syscall.EPIPE):
				return totalWritten, WireError{Kind: closedGracefully, Err: err}
			case errors.Is(err, syscall.ECONNRESET):
				return totalWritten, WireError{Kind: closedAbruptly, Err: err}
			case errors.As(err, &netErr):
				if netErr.Timeout() {
					return totalWritten, WireError{Kind: timeout, Err: err}
				}
				return totalWritten, WireError{Kind: unknown, Err: err}
			case errors.Is(err, os.ErrDeadlineExceeded):
				return totalWritten, WireError{Kind: timeout, Err: err}
			default:
				return totalWritten, WireError{Kind: unknown, Err: err}
			}
		}

		totalWritten += n
		writeAttempts++
	}

	// Successfully written all bytes
	return totalWritten, nil
}

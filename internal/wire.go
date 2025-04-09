// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package internal

type errKind int

const (
	unknown          errKind = 1
	timeout          errKind = 2
	closedGracefully errKind = 3
	closedAbruptly   errKind = 4
	corruptMessage   errKind = 5
	partialWrite     errKind = 6
)

type WireError struct {
	Kind errKind
	Err  error
}

func (e WireError) Error() string {
	return e.Err.Error()
}

func (e WireError) Unwrap() error {
	return e.Err
}

type Wire interface {
	Send([]byte) error
	Receive() ([]byte, error)
	Close()
}

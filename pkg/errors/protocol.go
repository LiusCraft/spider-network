package errors

import (
	"errors"
)

var (
	ErrToPacketTypeNotImplemented = errors.New("the protocol not implemented Packet")
)

package errors

import (
	"errors"
)

var (
	ErrToPacketTypeNotImplemented = errors.New("the protocol not implemented Packet")
	ErrCannotConvertType = errors.New("cannot convert this type")
)

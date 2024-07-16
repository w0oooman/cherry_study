package cherryError

import (
	"errors"
	"fmt"
)

type BaseError struct {
	Code    int32
	Message string
}

// NewError ctor
func NewError(err error, code int32) *BaseError {
	if err == nil {
		return &BaseError{
			Code: code,
		}
	}
	if err, ok := err.(*BaseError); ok {
		return err
	}
	e := &BaseError{
		Code:    code,
		Message: err.Error(),
	}
	return e

}

func (e *BaseError) Error() string {
	return e.Message
}

func (e *BaseError) ErrorCode() int32 {
	return e.Code
}

// CodeFromError returns the code of error.
// If error is nil, return empty string.
// If error is not a base error, returns unkown code
func CodeFromError(err error) int32 {
	if err == nil {
		return UnknownCode
	}

	baseError, ok := err.(*BaseError)
	if !ok {
		return UnknownCode
	}

	if baseError == nil {
		return UnknownCode
	}

	return baseError.ErrorCode()
}

func Error(text string) error {
	return errors.New(text)
}

func Errorf(format string, a ...interface{}) error {
	return errors.New(fmt.Sprintf(format, a...))
}

func Wrap(err error, text string) error {
	return Errorf("err:%v, text:%s", err, text)
}

func Wrapf(err error, format string, a ...interface{}) error {
	text := fmt.Sprintf(format, a...)
	return Wrap(err, text)
}

// session
var (
	SessionMemberNotFound    = Error("member not found in the group")
	SessionClosedGroup       = Error("group is closed")
	SessionDuplication       = Error("session has existed in the current group")
	SessionNotFoundInContext = Error("session not found in context")
)

// route
var (
	RouteFieldCantEmpty = Error("route field can not be empty")
	RouteInvalid        = Error("invalid route")
)

// packet
var (
	PacketWrongType              = Error("wrong packet type")
	PacketSizeExceed             = Error("codec: packet size exceed")
	PacketConnectClosed          = Error("client connection closed")
	PacketInvalidHeader          = Error("invalid header")
	PacketMsgSmallerThanExpected = Error("received less data than expected, EOF?")
	PacketHeadFuncNoSet          = Error("head func no set")
)

// message
var (
	ErrUnknownCode          = Error("unknown code")
	MessageWrongType        = Error("wrong message type")
	MessageInvalid          = Error("invalid message")
	MessageRouteNotFound    = Error("route info not found in dictionary")
	ErrReplyShouldBeNotNull = errors.New("reply must not be null")
)

var (
	ProtobufWrongValueType = Error("convert on wrong type value")
)

var (
	DiscoveryMemberListIsEmpty = Error("get member list is empty.")
)

// cluster
var (
	ClusterRPCClientIsStop = Error("rpc client is stop")
	ClusterNoImplement     = Error("no implement")
	NodeTypeIsNil          = Error("node type is nil.")
)

var (
	ActorPathError = Error("actor path is error.")
)

var (
	FuncIsNil     = Error("function is nil")
	FuncTypeError = Error("Is not func type")
)

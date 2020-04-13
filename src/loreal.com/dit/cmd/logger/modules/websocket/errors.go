package websocket

import "errors"

//ErrInvalidParameter - invalid parameter
var ErrInvalidParameter = errors.New("Invalid Parameter")

//ErrInvalidState - invalid state
var ErrInvalidState = errors.New("Invalid State")

//ErrInvalidAction - invalid action
var ErrInvalidAction = errors.New("Invalid Action")

//ErrOperationFailed - failed to process
var ErrOperationFailed = errors.New("Operation Failed")

//ErrPermissionDenied - Permission Denied!
var ErrPermissionDenied = errors.New("Permission Denied")

//ErrInvalidOwner - Invalid Owner!
var ErrInvalidOwner = errors.New("Invalid Owner")

// ErrReceiveTimeOut Receive notification timeout
var ErrReceiveTimeOut = errors.New("Receive notification timeout")

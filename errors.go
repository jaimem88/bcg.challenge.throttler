package throttler

import (
	"fmt"
	"net/http"
)

var (
	errInternalServerError = &Error{Code: http.StatusInternalServerError, Message: "Something went wrong :("}
	errMissingAccessToken  = &Error{Code: http.StatusUnauthorized, Message: "Missing access token"}
)

// Error describes custom error that can be used for logging and to write the response inside the handler
type Error struct {
	message string // msg used for logging purposes
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// Error implements error interface
func (e *Error) Error() string {
	return fmt.Sprintf("code: %d message: %s", e.Code, e.Message)
}

func (e *Error) msg(m string) *Error {
	e.message = m
	return e
}

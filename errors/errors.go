package errors

import (
	"fmt"
	"net/http"
)

// HTTP implements error and keeps track of http return codes.
type HTTP struct {
	error
	Message string
	Code    int
}

func (e HTTP) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// ToHTTP wraps the type assertion to change an error into an HTTP.
func ToHTTP(err error) *HTTP {
	if err == nil {
		return nil
	}
	rerr := &HTTP{
		Message: err.Error(),
		Code:    http.StatusInternalServerError,
	}
	if e, ok := err.(HTTP); ok {
		rerr.Code = e.Code
	}
	return rerr
}

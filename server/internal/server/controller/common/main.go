package common

import (
	"net/http"
)

type _Handler func(http.ResponseWriter, *http.Request) error

type ControllerMapValue struct {
	Handler   _Handler
	Protected bool
}

func NewCMV(handler _Handler, protected bool) *ControllerMapValue {
	return &ControllerMapValue{
		Handler:   handler,
		Protected: protected,
	}
}

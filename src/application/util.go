package application

import "fmt"

// note: these are internal requests, not http requests

type RequestHandlerFunc func(args ...interface{}) (interface{}, error)

type RequestHandlerMap map[string]RequestHandlerFunc

func (m RequestHandlerMap) Handle(ctx Context, cmd string, args ...interface{}) (interface{}, error) {
	fn, exists := m[cmd]
	if !exists {
		return nil, fmt.Errorf("config: command not recognized: %s", cmd)
	}
	return fn(args...)
}

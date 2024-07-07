package server

import (
	"fmt"
	"os"
)

var (
// TODO make errors
)

//go:generate mockery --name Logger
type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

//go:generate mockery --name Application
type Application interface {
	// TODO make interface
}

func Exitfail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

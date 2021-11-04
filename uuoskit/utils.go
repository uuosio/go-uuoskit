package uuoskit

import (
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

func DecodeHash256(hash string) ([]byte, error) {
	_hash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, newError(err)
	}
	if len(_hash) != 32 {
		return nil, newErrorf("invalid hash")
	}
	return _hash, nil
}

type DebugError struct {
	lines []string
	err   error
}

func (e *DebugError) Error() string {
	if DEBUG {
		errMsg := strings.Join(e.lines, "\n")
		return errMsg + "\n" + e.err.Error()
	} else {
		return e.err.Error()
	}
}

func NewDebugError(lineInfo string, err error) error {
	return &DebugError{[]string{lineInfo}, err}
}

func newError(err error) error {
	debugErr, ok := err.(*DebugError)
	if ok {
		if len(debugErr.lines) > 1000 {
			panic(err)
		}
		pc, fn, line, _ := runtime.Caller(1)
		lineInfo := fmt.Sprintf("%s[%s:%d]", runtime.FuncForPC(pc).Name(), fn, line)
		//insert lineInfo at the beginning of lines
		debugErr.lines = append(debugErr.lines[:1], debugErr.lines[0:]...)
		debugErr.lines[0] = lineInfo
		return debugErr
	} else {
		pc, fn, line, _ := runtime.Caller(1)
		lineInfo := fmt.Sprintf("[error] in %s[%s:%d]", runtime.FuncForPC(pc).Name(), fn, line)
		return NewDebugError(lineInfo, err)
	}
}

func newErrorf(format string, args ...interface{}) error {
	errMsg := fmt.Sprintf(format, args...)
	pc, fn, line, _ := runtime.Caller(1)
	lineInfo := fmt.Sprintf("[error] in %s[%s:%d]", runtime.FuncForPC(pc).Name(), fn, line)
	return NewDebugError(lineInfo, errors.New(errMsg))
}

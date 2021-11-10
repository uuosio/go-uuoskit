package uuoskit

import (
	"encoding/hex"
	"errors"
	"fmt"

	traceable_errors "github.com/go-errors/errors"
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

func newError(err error) error {
	if DEBUG {
		return traceable_errors.New(err.Error())
	} else {
		return err
	}
}

func newErrorf(format string, args ...interface{}) error {
	errMsg := fmt.Sprintf(format, args...)
	if DEBUG {
		return traceable_errors.New(errMsg)
	} else {
		return errors.New(errMsg)
	}
}

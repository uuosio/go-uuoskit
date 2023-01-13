package uuoskit

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	traceable_errors "github.com/go-errors/errors"
)

var (
	DEBUG = true
)

func SetDebug(debug bool) {
	DEBUG = debug
}

func GetDebug() bool {
	return DEBUG
}

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
		if _, ok := err.(*traceable_errors.Error); ok {
			return err
		}
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

func reverseBytes(aa []byte) {
	for i, j := 0, len(aa)-1; i < j; i, j = i+1, j-1 {
		aa[i], aa[j] = aa[j], aa[i]
	}
}

func StringToInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, newError(err)
	}
	return i, nil
}

func StripString(v string) (string, bool) {
	v, err := strconv.Unquote(v)
	if err != nil {
		return "", false
	}
	return v, true
}

func IsSymbolValid(sym string) bool {
	if len(sym) > 7 || len(sym) <= 0 {
		return false
	}
	_vv := []byte(sym)
	i := 0
	for ; i < len(_vv); i++ {
		c := _vv[i]
		if c >= byte('A') && c <= byte('Z') {
		} else if c == 0 {
			for ; i < len(_vv); i++ {
				if _vv[i] != 0 {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func ParseAsset(v string) ([]byte, bool) {
	vv := strings.Split(v, " ")
	if len(vv) != 2 {
		return nil, false
	}
	amount := vv[0]
	symbol_code := vv[1]
	if !IsSymbolValid(symbol_code) {
		return nil, false
	}
	_, err := strconv.ParseFloat(amount, 64)

	if err != nil {
		return nil, false
	}
	_amount := strings.Split(amount, ".")
	if len(_amount) > 2 {
		return nil, false
	}

	precision := 0
	if len(_amount) == 2 {
		precision = len(_amount[1])
	}

	amount = strings.Replace(amount, ".", "", 1)
	nAmount, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return nil, false
	}

	if nAmount < 0 || nAmount > math.MaxInt64 {
		return nil, false
	}

	enc := NewEncoder(8)
	enc.PackInt64(nAmount)

	enc.WriteBytes([]byte{byte(precision)})
	_symbol_code := make([]byte, 7)
	copy(_symbol_code[:], []byte(symbol_code)[:])
	enc.WriteBytes(_symbol_code)
	return enc.GetBytes(), true
}

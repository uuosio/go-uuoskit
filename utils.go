package main

import (
	"encoding/hex"
	"errors"
)

func DecodeHash256(hash string) ([]byte, error) {
	_hash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	if len(_hash) != 32 {
		return nil, errors.New("invalid hash")
	}
	return _hash, nil
}

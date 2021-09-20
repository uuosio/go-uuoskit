package main

import (
	"encoding/hex"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// "bool"
// "int8"
// "uint8"
// "int16"
// "uint16"
// "int32"
// "uint32"
// "int64"
// "uint64"
// "int128"
// "uint128"
// "varint32"
// "varuint32"

// // TODO: Add proper support for floating point types. For now this is good enough.
// "float32"
// "float64"
// "float128"

// "time_point"
// "time_point_sec"
// "block_timestamp_type"

// "name"

// "bytes"
// "string"

// "checksum160"
// "checksum256"
// "checksum512"

// "public_key"
// "signature"

// "symbol"
// "symbol_code"
// "asset"
// "extended_asset"

var gAbi = `{
	"version": "eosio::abi/1.0",
	"types": [
	],
	"structs": [
			{
			"name": "test",
			"base": "",
			"fields": [
				{
					"name": "t",
					"type": "%s"
				}
			]
		}
	],
	"actions": [{
		"name": "test",
		"type": "test",
		"ricardian_contract": ""
	}],
	"tables": [],
	"ricardian_clauses": [],
	"error_messages": [],
	"abi_extensions": []
}
`

func AssertPackAbiValueError(t *testing.T, typ string, value string, expectedErr error) {
	assert := assert.New(t)
	_abi := fmt.Sprintf(gAbi, typ)
	s := GetABISerializer()
	s.SetContractABI("test", []byte(_abi))

	_, err := s.PackAbiStructByName("test", "test", fmt.Sprintf(`{"t": %s}`, value))
	assert.NotEqual(err, nil, "expected error")
	// t.Log(expectedErr, err)
	assert.Equal(expectedErr.Error(), err.Error(), "bad error")
}

func AssertPackAbiValue(t *testing.T, typ string, value string, expected string) {
	_abi := fmt.Sprintf(gAbi, typ)
	s := GetABISerializer()
	s.SetContractABI("test", []byte(_abi))

	r, err := s.PackAbiStructByName("test", "test", fmt.Sprintf(`{"t": %s}`, value))
	if err != nil {
		panic(err)
	}
	assert := assert.New(t)
	assert.Equal(expected, hex.EncodeToString(r), "bad packed value")
}

func TestPackAbiType(t *testing.T) {
	AssertPackAbiValue(t, "bool", "true", "01")
	AssertPackAbiValue(t, "bool", "false", "00")
	// "bool"
	// "int8"
	AssertPackAbiValueError(t, "int8", "128", fmt.Errorf("int8 overflow: 128"))
	AssertPackAbiValueError(t, "int8", "255", fmt.Errorf("int8 overflow: 255"))
	AssertPackAbiValue(t, "int8", "127", "7f")
	AssertPackAbiValue(t, "int8", "-1", "ff")
	AssertPackAbiValue(t, "int8", "0", "00")
	// "uint8"
	AssertPackAbiValue(t, "uint8", "255", "ff")
	AssertPackAbiValue(t, "uint8", "0", "00")

	// "int16"
	AssertPackAbiValueError(t, "int16", "32768", fmt.Errorf("int16 overflow: 32768"))
	AssertPackAbiValue(t, "int16", "32767", "ff7f")
	AssertPackAbiValue(t, "int16", "0", "0000")
	AssertPackAbiValue(t, "int16", "-1", "ffff")
	// "uint16"
	AssertPackAbiValue(t, "uint16", "65535", "ffff")
	AssertPackAbiValue(t, "uint16", "0", "0000")
	// "int32"
	AssertPackAbiValue(t, "int32", "2147483647", "ffffff7f")
	AssertPackAbiValue(t, "int32", "0", "00000000")
	AssertPackAbiValue(t, "int32", "-1", "ffffffff")
	// "uint32"
	AssertPackAbiValue(t, "uint32", "4294967295", "ffffffff")
	AssertPackAbiValue(t, "uint32", "0", "00000000")
	// "int64"
	AssertPackAbiValue(t, "int64", "9223372036854775807", "ffffffffffffff7f")
	AssertPackAbiValue(t, "int64", "0", "0000000000000000")
	AssertPackAbiValue(t, "int64", "-1", "ffffffffffffffff")
	// "uint64"
	AssertPackAbiValue(t, "uint64", "18446744073709551615", "ffffffffffffffff")
	AssertPackAbiValue(t, "uint64", "0", "0000000000000000")
	// "int128"
	AssertPackAbiValue(t, "int128", `"0x70680300000000000000000000000000"`, "70680300000000000000000000000000")
	// "uint128"
	AssertPackAbiValue(t, "uint128", `"0x70680300000000000000000000000000"`, "70680300000000000000000000000000")
	// "varint32"
	AssertPackAbiValue(t, "varint32", "128", "8002")
	// "varuint32"
	AssertPackAbiValue(t, "varuint32", "128", "8001")

	// "float32"
	AssertPackAbiValue(t, "float32", "11.22", "1f853341")
	// "float64"
	AssertPackAbiValue(t, "float32", "-11.22", "1f8533c1")

	// "float128"
	AssertPackAbiValue(t, "float128", `"0x70680300000000000000000000000000"`, "70680300000000000000000000000000")

	// "time_point"
	// "time_point_sec"
	// "block_timestamp_type"

	// "name"
	AssertPackAbiValue(t, "name", `"hello"`, "00000000001aa36a")

	// "bytes"
	AssertPackAbiValue(t, "bytes", `"ff8000"`, "03ff8000")
	// "string"
	AssertPackAbiValue(t, "string", `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`, "80016161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161")

	// "checksum160"
	AssertPackAbiValue(t, "checksum160", `"aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb"`, "aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb")

	// "checksum256"
	AssertPackAbiValue(t, "checksum256", `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"`, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	// "checksum512"
	AssertPackAbiValue(t, "checksum512", `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"`, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	// "public_key"
	// "signature"

	// "symbol"
	// "symbol_code"
	// "asset"
	// "extended_asset"

	// t.Log(hex.EncodeToString(r))
}

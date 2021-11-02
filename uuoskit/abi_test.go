package uuoskit

import (
	"encoding/hex"
	"io/ioutil"

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

func TestBuyRam(t *testing.T) {
	//read abi from file
	abi, err := ioutil.ReadFile("./data/eosio.system.abi")
	if err != nil {
		t.Error(err)
	}
	s := GetABISerializer()
	s.SetContractABI("eosio", abi)

	args := `{"payer": "hello", "receiver": "mixincrossss", "bytes": 10485760}`
	r, err := s.PackAbiStructByName("eosio", "buyrambytes", args)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%x", r)

	r, err = s.PackActionArgs("eosio", "buyrambytes", []byte(args))
	if err != nil {
		t.Error(err)
	}
	t.Logf("%x", r)
}

func TestAbiTypeDefine(t *testing.T) {
	abi := `{
	"version": "eosio::abi/1.0",
	"types": [
		{
		"new_type_name": "account_name",
		"type": "name"
		},
		{
		"new_type_name": "permission_name",
		"type": "name"
		},
		{
		"new_type_name": "action_name",
		"type": "name"
		},
		{
		"new_type_name": "table_name",
		"type": "name"
		},
		{
		"new_type_name": "transaction_id_type",
		"type": "checksum256"
		},
		{
		"new_type_name": "block_id_type",
		"type": "checksum256"
		},
		{
		"new_type_name": "weight_type",
		"type": "uint16"
		}
	],
	"structs": [
		{
		"name": "permission_level",
		"base": "",
		"fields": [
		{
		"name": "actor",
		"type": "account_name"
		},
		{
		"name": "permission",
		"type": "permission_name"
		}
		]
		},
		{
		"name": "action",
		"base": "",
		"fields": [
		{
		"name": "account",
		"type": "account_name"
		},
		{
		"name": "name",
		"type": "action_name"
		},
		{
		"name": "authorization",
		"type": "permission_level[]"
		},
		{
		"name": "data",
		"type": "bytes"
		}
		]
		},
		{
		"name": "extension",
		"base": "",
		"fields": [
		{
		"name": "type",
		"type": "uint16"
		},
		{
		"name": "data",
		"type": "bytes"
		}
		]
		},
		{
		"name": "transaction_header",
		"base": "",
		"fields": [
		{
		"name": "expiration",
		"type": "time_point_sec"
		},
		{
		"name": "ref_block_num",
		"type": "uint16"
		},
		{
		"name": "ref_block_prefix",
		"type": "uint32"
		},
		{
		"name": "max_net_usage_words",
		"type": "varuint32"
		},
		{
		"name": "max_cpu_usage_ms",
		"type": "uint8"
		},
		{
		"name": "delay_sec",
		"type": "varuint32"
		}
		]
		},
		{
		"name": "transaction",
		"base": "transaction_header",
		"fields": [
		{
		"name": "context_free_actions",
		"type": "action[]"
		},
		{
		"name": "actions",
		"type": "action[]"
		},
		{
		"name": "transaction_extensions",
		"type": "extension[]"
		}
		]
		},
		{
		"name": "producer_key",
		"base": "",
		"fields": [
		{
		"name": "producer_name",
		"type": "account_name"
		},
		{
		"name": "block_signing_key",
		"type": "public_key"
		}
		]
		},
		{
		"name": "producer_schedule",
		"base": "",
		"fields": [
		{
		"name": "version",
		"type": "uint32"
		},
		{
		"name": "producers",
		"type": "producer_key[]"
		}
		]
		},
		{
		"name": "block_header",
		"base": "",
		"fields": [
		{
		"name": "timestamp",
		"type": "uint32"
		},
		{
		"name": "producer",
		"type": "account_name"
		},
		{
		"name": "confirmed",
		"type": "uint16"
		},
		{
		"name": "previous",
		"type": "block_id_type"
		},
		{
		"name": "transaction_mroot",
		"type": "checksum256"
		},
		{
		"name": "action_mroot",
		"type": "checksum256"
		},
		{
		"name": "schedule_version",
		"type": "uint32"
		},
		{
		"name": "new_producers",
		"type": "producer_schedule?"
		},
		{
		"name": "header_extensions",
		"type": "extension[]"
		}
		]
		},
		{
		"name": "key_weight",
		"base": "",
		"fields": [
		{
		"name": "key",
		"type": "public_key"
		},
		{
		"name": "weight",
		"type": "weight_type"
		}
		]
		},
		{
		"name": "permission_level_weight",
		"base": "",
		"fields": [
		{
		"name": "permission",
		"type": "permission_level"
		},
		{
		"name": "weight",
		"type": "weight_type"
		}
		]
		},
		{
		"name": "wait_weight",
		"base": "",
		"fields": [
		{
		"name": "wait_sec",
		"type": "uint32"
		},
		{
		"name": "weight",
		"type": "weight_type"
		}
		]
		},
		{
		"name": "authority",
		"base": "",
		"fields": [
		{
		"name": "threshold",
		"type": "uint32"
		},
		{
		"name": "keys",
		"type": "key_weight[]"
		},
		{
		"name": "accounts",
		"type": "permission_level_weight[]"
		},
		{
		"name": "waits",
		"type": "wait_weight[]"
		}
		]
		},
		{
		"name": "newaccount",
		"base": "",
		"fields": [
		{
		"name": "creator",
		"type": "account_name"
		},
		{
		"name": "name",
		"type": "account_name"
		},
		{
		"name": "owner",
		"type": "authority"
		},
		{
		"name": "active",
		"type": "authority"
		}
		]
		},
		{
		"name": "setcode",
		"base": "",
		"fields": [
		{
		"name": "account",
		"type": "account_name"
		},
		{
		"name": "vmtype",
		"type": "uint8"
		},
		{
		"name": "vmversion",
		"type": "uint8"
		},
		{
		"name": "code",
		"type": "bytes"
		}
		]
		},
		{
		"name": "setabi",
		"base": "",
		"fields": [
		{
		"name": "account",
		"type": "account_name"
		},
		{
		"name": "abi",
		"type": "bytes"
		}
		]
		},
		{
		"name": "updateauth",
		"base": "",
		"fields": [
		{
		"name": "account",
		"type": "account_name"
		},
		{
		"name": "permission",
		"type": "permission_name"
		},
		{
		"name": "parent",
		"type": "permission_name"
		},
		{
		"name": "auth",
		"type": "authority"
		}
		]
		},
		{
		"name": "deleteauth",
		"base": "",
		"fields": [
		{
		"name": "account",
		"type": "account_name"
		},
		{
		"name": "permission",
		"type": "permission_name"
		}
		]
		},
		{
		"name": "linkauth",
		"base": "",
		"fields": [
		{
		"name": "account",
		"type": "account_name"
		},
		{
		"name": "code",
		"type": "account_name"
		},
		{
		"name": "type",
		"type": "action_name"
		},
		{
		"name": "requirement",
		"type": "permission_name"
		}
		]
		},
		{
		"name": "unlinkauth",
		"base": "",
		"fields": [
		{
		"name": "account",
		"type": "account_name"
		},
		{
		"name": "code",
		"type": "account_name"
		},
		{
		"name": "type",
		"type": "action_name"
		}
		]
		},
		{
		"name": "canceldelay",
		"base": "",
		"fields": [
		{
		"name": "canceling_auth",
		"type": "permission_level"
		},
		{
		"name": "trx_id",
		"type": "transaction_id_type"
		}
		]
		},
		{
		"name": "onerror",
		"base": "",
		"fields": [
		{
		"name": "sender_id",
		"type": "uint128"
		},
		{
		"name": "sent_trx",
		"type": "bytes"
		}
		]
		},
		{
		"name": "onblock",
		"base": "",
		"fields": [
		{
		"name": "header",
		"type": "block_header"
		}
		]
		}
	],
	"actions": [
		{
		"name": "newaccount",
		"type": "newaccount",
		"ricardian_contract": ""
		},
		{
		"name": "setcode",
		"type": "setcode",
		"ricardian_contract": ""
		},
		{
		"name": "setabi",
		"type": "setabi",
		"ricardian_contract": ""
		},
		{
		"name": "updateauth",
		"type": "updateauth",
		"ricardian_contract": ""
		},
		{
		"name": "deleteauth",
		"type": "deleteauth",
		"ricardian_contract": ""
		},
		{
		"name": "linkauth",
		"type": "linkauth",
		"ricardian_contract": ""
		},
		{
		"name": "unlinkauth",
		"type": "unlinkauth",
		"ricardian_contract": ""
		},
		{
		"name": "canceldelay",
		"type": "canceldelay",
		"ricardian_contract": ""
		},
		{
		"name": "onerror",
		"type": "onerror",
		"ricardian_contract": ""
		},
		{
		"name": "onblock",
		"type": "onblock",
		"ricardian_contract": ""
		}
	],
	"tables": [],
	"ricardian_clauses": [],
	"error_messages": [],
	"abi_extensions": [],
	"variants": [],
	"action_results": [],
	"kv_tables": {}
	}`
	err := GetABISerializer().SetContractABI("eosio", []byte(abi))
	if err != nil {
		t.Error(err)
	}
	args := `{"account":"eosio", "vmtype":0, "vmversion":0, "code": "1122"}`
	r, err := GetABISerializer().PackActionArgs("eosio", "setcode", []byte(args))
	if err != nil {
		t.Error(err)
	}
	t.Log(hex.EncodeToString(r))
}

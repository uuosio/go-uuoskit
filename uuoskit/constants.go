package uuoskit

var eosioTokenAbi = `
{
    "____comment": "This file was generated with eosio-abigen. DO NOT EDIT ",
    "version": "eosio::abi/1.1",
    "types": [],
    "structs": [
        {
            "name": "account",
            "base": "",
            "fields": [
                {
                    "name": "balance",
                    "type": "asset"
                }
            ]
        },
        {
            "name": "close",
            "base": "",
            "fields": [
                {
                    "name": "owner",
                    "type": "name"
                },
                {
                    "name": "symbol",
                    "type": "symbol"
                }
            ]
        },
        {
            "name": "create",
            "base": "",
            "fields": [
                {
                    "name": "issuer",
                    "type": "name"
                },
                {
                    "name": "maximum_supply",
                    "type": "asset"
                }
            ]
        },
        {
            "name": "currency_stats",
            "base": "",
            "fields": [
                {
                    "name": "supply",
                    "type": "asset"
                },
                {
                    "name": "max_supply",
                    "type": "asset"
                },
                {
                    "name": "issuer",
                    "type": "name"
                }
            ]
        },
        {
            "name": "issue",
            "base": "",
            "fields": [
                {
                    "name": "to",
                    "type": "name"
                },
                {
                    "name": "quantity",
                    "type": "asset"
                },
                {
                    "name": "memo",
                    "type": "string"
                }
            ]
        },
        {
            "name": "open",
            "base": "",
            "fields": [
                {
                    "name": "owner",
                    "type": "name"
                },
                {
                    "name": "symbol",
                    "type": "symbol"
                },
                {
                    "name": "ram_payer",
                    "type": "name"
                }
            ]
        },
        {
            "name": "retire",
            "base": "",
            "fields": [
                {
                    "name": "quantity",
                    "type": "asset"
                },
                {
                    "name": "memo",
                    "type": "string"
                }
            ]
        },
        {
            "name": "transfer",
            "base": "",
            "fields": [
                {
                    "name": "from",
                    "type": "name"
                },
                {
                    "name": "to",
                    "type": "name"
                },
                {
                    "name": "quantity",
                    "type": "asset"
                },
                {
                    "name": "memo",
                    "type": "string"
                }
            ]
        }
    ],
    "actions": [
        {
            "name": "close",
            "type": "close"
        },
        {
            "name": "create",
            "type": "create"
        },
        {
            "name": "issue",
            "type": "issue"
        },
        {
            "name": "open",
            "type": "open"
        },
        {
            "name": "retire",
            "type": "retire"
        },
        {
            "name": "transfer",
            "type": "transfer"
        }
    ],
    "tables": [
        {
            "name": "accounts",
            "type": "account",
            "index_type": "i64",
            "key_names": [],
            "key_types": []
        },
        {
            "name": "stat",
            "type": "currency_stats",
            "index_type": "i64",
            "key_names": [],
            "key_types": []
        }
    ],
    "ricardian_clauses": [],
    "variants": []
}
`

var gBaseTypes = map[string]bool{
	"bool":                 true,
	"int8":                 true,
	"uint8":                true,
	"int16":                true,
	"uint16":               true,
	"int32":                true,
	"uint32":               true,
	"int64":                true,
	"uint64":               true,
	"int128":               true,
	"uint128":              true,
	"varint32":             true,
	"varuint32":            true,
	"float32":              true,
	"float64":              true,
	"float128":             true,
	"time_point":           true,
	"time_point_sec":       true,
	"block_timestamp_type": true,
	"name":                 true,
	"bytes":                true,
	"string":               true,
	"checksum160":          true,
	"checksum256":          true,
	"checksum512":          true,
	"public_key":           true,
	"signature":            true,
	"symbol":               true,
	"symbol_code":          true,
	"asset":                true,
	"extended_asset":       true,
}

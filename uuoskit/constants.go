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

var baseTypes = []string{
	"bool",
	"int8",
	"uint8",
	"int16",
	"uint16",
	"int32",
	"uint32",
	"int64",
	"uint64",
	"int128",
	"uint128",
	"varint32",
	"varuint32",
	"float32",
	"float64",
	"float128",
	"time_point",
	"time_point_sec",
	"block_timestamp_type",
	"name",
	"bytes",
	"string",
	"checksum160",
	"checksum256",
	"checksum512",
	"public_key",
	"signature",
	"symbol",
	"symbol_code",
	"asset",
	"extended_asset",
}

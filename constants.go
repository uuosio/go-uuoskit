package main

var baseABI = `
{
    "version": "eosio::abi/1.1",
    "types": [],
    "structs": [
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
        }
    ],
    "actions": [
    ],
    "tables": [

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

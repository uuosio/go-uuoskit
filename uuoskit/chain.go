package uuoskit

import "encoding/json"

// {'server_version': '6d383cb1',
//  'chain_id': '9b1605a3f7f14995641c6b19413841c26ca86747f054241951a298b556160674',
//  'head_block_num': 5918917,
//  'last_irreversible_block_num': 5918916,
//  'last_irreversible_block_id': '005a50c451107fd4d94493f152d832a6420aa7945d51974dca56b2a1f3dfe5fe',
//  'head_block_id': '005a50c54ce7a3c4859681ebd3aea624befefaa219c70391fb43ac9453d9cdce',
//  'head_block_time': '2021-09-01T06:27:45.000',
//  'head_block_producer': 'eosio',
//  'virtual_block_cpu_limit': 1500000000,
//  'virtual_block_net_limit': 1048576000,
//  'block_cpu_limit': 1500000,
//  'block_net_limit': 1048576,
//  'server_version_string': 'v2.1.0-rc1',
//  'fork_db_head_block_num': 5918917,
//  'fork_db_head_block_id': '005a50c54ce7a3c4859681ebd3aea624befefaa219c70391fb43ac9453d9cdce',
//  'server_full_version_string': 'v2.1.0-rc1-2afb85e3cca39054ae690219cf4751ebd309df67-dirty',
//  'last_irreversible_block_time': '2021-09-01T06:27:42.000'}

type ChainInfo struct {
	ServerVersion             string `json:"server_version"`
	ChainID                   string `json:"chain_id"`
	HeadBlockNum              int64  `json:"head_block_num"`
	LastIrreversibleBlockNum  int64  `json:"last_irreversible_block_num"`
	LastIrreversibleBlockID   string `json:"last_irreversible_block_id"`
	HeadBlockID               string `json:"head_block_id"`
	HeadBlockTime             string `json:"head_block_time"`
	HeadBlockProducer         string `json:"head_block_producer"`
	VirtualBlockCPULimit      int64  `json:"virtual_block_cpu_limit"`
	VirtualBlockNetLimit      int64  `json:"virtual_block_net_limit"`
	BlockCPULimit             int64  `json:"block_cpu_limit"`
	BlockNetLimit             int64  `json:"block_net_limit"`
	ServerVersionString       string `json:"server_version_string"`
	ForkDBHeadBlockNum        int64  `json:"fork_db_head_block_num"`
	ForkDBHeadBlockID         string `json:"fork_db_head_block_id"`
	ServerFullVersionString   string `json:"server_full_version_string"`
	LastIrreversibleBlockTime string `json:"last_irreversible_block_time"`
}

func NewChainInfo(info []byte) (*ChainInfo, error) {
	chainInfo := &ChainInfo{}
	if err := json.Unmarshal(info, chainInfo); err != nil {
		return nil, newError(err)
	}
	return chainInfo, nil
}

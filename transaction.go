package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/learnforpractice/go-secp256k1/secp256k1"
)

type TransactionExtension struct {
	Type uint16
	Data []byte
}

func (a *TransactionExtension) Size() int {
	return 2 + 5 + len(a.Data)
}

func (t *TransactionExtension) Pack() []byte {
	enc := NewEncoder(2 + 5 + len(t.Data))
	enc.Pack(t.Type)
	enc.Pack(t.Data)
	return enc.GetBytes()
}

func (t *TransactionExtension) Unpack(data []byte) (int, error) {
	var err error
	dec := NewDecoder(data)
	t.Type, err = dec.UnpackUint16()
	if err != nil {
		return 0, err
	}

	t.Data, err = dec.UnpackBytes()
	if err != nil {
		return 0, err
	}
	return dec.Pos(), nil
}

type Transaction struct {
	// time_point_sec  expiration;
	// uint16_t        ref_block_num;
	// uint32_t        ref_block_prefix;
	// unsigned_int    max_net_usage_words = 0UL; /// number of 8 byte words this transaction can serialize into after compressions
	// uint8_t         max_cpu_usage_ms = 0UL; /// number of CPU usage units to bill transaction for
	// unsigned_int    delay_sec = 0UL; /// number of seconds to delay transaction, default: 0
	Expiration     uint32
	RefBlockNum    uint16
	RefBlockPrefix uint32
	//[VLQ or Base-128 encoding](https://en.wikipedia.org/wiki/Variable-length_quantity)
	//unsigned_int vaint (eosio.cdt/libraries/eosiolib/core/eosio/varint.hpp)
	MaxNetUsageWords   uint32
	MaxCpuUsageMs      uint8
	DelaySec           uint32 //unsigned_int
	ContextFreeActions []Action
	Actions            []Action
	Extention          []TransactionExtension
}

// '{"signatures":
// ["SIG_K1_KbSF8BCNVA95KzR1qLmdn4VnxRoLVFQ1fZ8VV5gVdW1hLfGBdcwEc93hF7FBkWZip1tq2Ps27UZxceaR3hYwAjKL7j59q8"],
// "compression":"none",
// "packed_context_free_data":"",
// "packed_trx":"4bc52d61a9dd120d4ade000000000100a6823403ea3055000000572d3ccdcd0110428a97721aa36a00000000a8ed32323b10428a97721aa36a0000000000000e3d102700000000000004454f53000000001a7472616e736665722066726f6d20616c69636520746f20626f6200"}'

type TxBytes []byte

func (m TxBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(m))
}

type PackedTransaction struct {
	chainId       [32]byte
	tx            *Transaction
	Signatures    []string `json:"signatures"`
	Compression   string   `json:"compression"`
	PackedContext TxBytes  `json:"packed_context_free_data"`
	PackedTx      TxBytes  `json:"packed_trx"`
}

func NewTransaction(expiration int) *Transaction {
	t := &Transaction{}
	t.Expiration = uint32(expiration)
	// t.RefBlockNum = uint16(taposBlockNum)
	// t.RefBlockPrefix = uint32(taposBlockPrefix)
	t.MaxNetUsageWords = uint32(0)
	t.MaxCpuUsageMs = uint8(0)
	// t.DelaySec = uint32(delaySec)
	return t
}

func GetRefBlockNum(refBlock []byte) uint32 {
	return uint32(refBlock[0])<<24 | uint32(refBlock[1])<<16 | uint32(refBlock[2])<<8 | uint32(refBlock[3])
}

func GetRefBlockPrefix(refBlock []byte) uint32 {
	return uint32(refBlock[11])<<24 | uint32(refBlock[10])<<16 | uint32(refBlock[9])<<8 | uint32(refBlock[8])
}

func (t *Transaction) SetReferenceBlock(refBlock string) error {
	id, err := hex.DecodeString(refBlock)
	if err != nil {
		return err
	}
	t.RefBlockNum = uint16(GetRefBlockNum(id))
	t.RefBlockPrefix = GetRefBlockPrefix(id)
	return nil
}

func (t *Transaction) AddAction(a *Action) {
	t.Actions = append(t.Actions, *a)
}

func (t *Transaction) Pack() []byte {
	initSize := 4 + 2 + 4 + 5 + 1 + 5

	initSize += 5 // Max varint size
	for _, action := range t.ContextFreeActions {
		initSize += action.Size()
	}

	initSize += 5 // Max varint size
	for _, action := range t.Actions {
		initSize += action.Size()
	}

	initSize += 5 // Max varint size
	for _, extention := range t.Extention {
		initSize += extention.Size()
	}
	enc := NewEncoder(initSize)
	enc.Pack(t.Expiration)
	enc.Pack(t.RefBlockNum)
	enc.Pack(t.RefBlockPrefix)
	enc.PackVarUint32(t.MaxNetUsageWords)
	enc.PackUint8(t.MaxCpuUsageMs)
	enc.PackVarUint32(t.DelaySec)

	enc.PackLength(len(t.ContextFreeActions))
	for _, action := range t.ContextFreeActions {
		enc.Pack(&action)
	}

	enc.PackLength(len(t.Actions))
	for _, action := range t.Actions {
		enc.Pack(&action)
	}

	enc.PackLength(len(t.Extention))
	for _, extention := range t.Extention {
		enc.Pack(&extention)
	}
	return enc.GetBytes()
}

func (t *Transaction) Unpack(data []byte) (int, error) {
	var err error

	dec := NewDecoder(data)
	t.Expiration, err = dec.UnpackUint32()
	if err != nil {
		return 0, err
	}

	t.RefBlockNum, err = dec.UnpackUint16()
	if err != nil {
		return 0, err
	}

	t.RefBlockPrefix, err = dec.UnpackUint32()
	if err != nil {
		return 0, err
	}

	t.MaxNetUsageWords, err = dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	t.MaxCpuUsageMs, err = dec.UnpackUint8()
	if err != nil {
		return 0, err
	}

	t.DelaySec, err = dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	contextFreeActionLength, err := dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	t.ContextFreeActions = make([]Action, contextFreeActionLength)
	for i := 0; i < int(contextFreeActionLength); i++ {
		_, err := dec.Unpack(&t.ContextFreeActions[i])
		if err != nil {
			return 0, err
		}
	}

	actionLength, err := dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	t.Actions = make([]Action, actionLength)
	for i := 0; i < int(actionLength); i++ {
		_, err := dec.Unpack(&t.Actions[i])
		if err != nil {
			return 0, err
		}
	}

	extentionLength, err := dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}
	t.Extention = make([]TransactionExtension, extentionLength)
	for i := 0; i < int(extentionLength); i++ {
		t.Extention[i].Type, err = dec.UnpackUint16()
		if err != nil {
			return 0, err
		}

		t.Extention[i].Data, err = dec.UnpackBytes()
		if err != nil {
			return 0, err
		}
	}
	return dec.Pos(), nil
}

func (t *Transaction) Sign(privKey string, chainId string) (string, error) {
	data := t.Pack()

	hash := sha256.New()
	hash.Write(data)
	digest := hash.Sum(nil)

	priv, err := secp256k1.NewPrivateKeyFromBase58(privKey)
	if err != nil {
		return "", err
	}
	sign, err := secp256k1.Sign(digest, priv)
	if err != nil {
		return "", err
	}
	// type PackedTransaction struct {
	// 	Signatures    []string `json:"signatures"`
	// 	Compression   string   `json:"compression"`
	// 	PackedContext []byte   `json:"packed_context_free_data"`
	// 	PackedTx     []byte   `json:"packed_trx"`
	// }
	return sign.String(), nil
}

func NewPackedtransaction(tx *Transaction) *PackedTransaction {
	packed := &PackedTransaction{}
	packed.Compression = "none"
	packed.PackedTx = nil
	packed.tx = tx
	return packed
}

//SetChainId
func (t *PackedTransaction) SetChainId(chainId string) error {
	id, err := DecodeHash256(chainId)
	if err != nil {
		return err
	}
	copy(t.chainId[:], id)
	return nil
}

func (t *PackedTransaction) sign(priv *secp256k1.PrivateKey, chainId []byte) error {
	if t.PackedTx == nil {
		t.PackedTx = t.tx.Pack()
	}

	hash := sha256.New()
	hash.Write(chainId)
	hash.Write(t.PackedTx)
	//TODO: hash context_free_data
	cfdHash := [32]byte{}
	hash.Write(cfdHash[:])
	digest := hash.Sum(nil)

	sign, err := priv.Sign(digest)
	if err != nil {
		return err
	}

	newSign := sign.String()
	for i := range t.Signatures {
		sig := t.Signatures[i]
		if sig == newSign {
			return nil
		}
	}

	t.Signatures = append(t.Signatures, sign.String())
	return nil
}

func (t *PackedTransaction) Sign(pubKey string) error {
	priv, err := GetWallet().GetPrivateKey(pubKey)
	if err != nil {
		return err
	}
	empty := false
	for i := 0; i < 32; i++ {
		if t.chainId[i] != 0 {
			empty = false
			break
		}
	}

	if empty {
		return errors.New("chainId is empty")
	}

	return t.sign(priv, t.chainId[:])
}

func (t *PackedTransaction) SignByPrivateKey(privKey string, chainId string) error {
	priv, err := secp256k1.NewPrivateKeyFromBase58(privKey)
	if err != nil {
		return err
	}

	_chainId, err := DecodeHash256(chainId)
	if err != nil {
		return err
	}
	return t.sign(priv, _chainId)
}

func (t *PackedTransaction) String() string {
	packed, _ := json.Marshal(t)
	return string(packed)
}

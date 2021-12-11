package main

//#include <stdint.h>
//#include <stddef.h>
//#include <stdbool.h>
// static char* get_p(char **pp, int i)
// {
//	    return pp[i];
// }
import "C"

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"runtime"
	"unsafe"

	traceable_errors "github.com/go-errors/errors"
	secp256k1 "github.com/uuosio/go-secp256k1"
	"github.com/uuosio/go-uuoskit/uuoskit"
)

func renderData(data interface{}) *C.char {
	ret := map[string]interface{}{"data": data}
	result, _ := json.Marshal(ret)
	return C.CString(string(result))
}

func renderError(err error) *C.char {
	if _err, ok := err.(*traceable_errors.Error); ok {
		errMsg := _err.ErrorStack()
		ret := map[string]interface{}{"error": errMsg}
		result, _ := json.Marshal(ret)
		return C.CString(string(result))
	} else {
		pc, fn, line, _ := runtime.Caller(1)
		errMsg := fmt.Sprintf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
		ret := map[string]interface{}{"error": errMsg}
		result, _ := json.Marshal(ret)
		return C.CString(string(result))
	}
}

//export init_
func init_() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	secp256k1.Init()
	if nil == uuoskit.GetABISerializer() {
		panic("abi serializer not initialized")
	}
}

//export say_hello_
func say_hello_(name *C.char) {
	_name := C.GoString(name)
	log.Println("hello,world", _name)
}

//export wallet_import_
func wallet_import_(name *C.char, priv *C.char) *C.char {
	_name := C.GoString(name)
	_priv := C.GoString(priv)
	err := uuoskit.GetWallet().Import(_name, _priv)
	if err != nil {
		return renderError(err)
	}
	// log.Println("import", _name, _priv)
	return renderData("ok")
}

//export wallet_get_public_keys_
func wallet_get_public_keys_() *C.char {
	keys := uuoskit.GetWallet().GetPublicKeys()
	return renderData(keys)
}

//export wallet_sign_digest_
func wallet_sign_digest_(digest *C.char, pubKey *C.char) *C.char {
	_pubKey := C.GoString(pubKey)
	// log.Println("++++++++wallet_sign_digest_:", C.GoString(digest))
	_digest, err := hex.DecodeString(C.GoString(digest))
	if err != nil {
		return renderError(err)
	}

	sign, err := uuoskit.GetWallet().Sign(_digest, _pubKey)
	if err != nil {
		return renderError(err)
	}
	return renderData(sign.String())
}

var gPackedTxs []*uuoskit.PackedTransaction

func validateIndex(idx C.int64_t) error {
	if idx < 0 || idx >= C.int64_t(len(gPackedTxs)) {
		return fmt.Errorf("invalid idx")
	}

	if gPackedTxs[idx] == nil {
		return fmt.Errorf("invalid idx")
	}

	return nil
}

func addPackedTx(packedTx *uuoskit.PackedTransaction) C.int64_t {
	if gPackedTxs == nil {
		gPackedTxs = make([]*uuoskit.PackedTransaction, 0, 10)
	}

	if len(gPackedTxs) >= 1024 {
		return C.int64_t(-1)
	}

	for i := 0; i < len(gPackedTxs); i++ {
		if gPackedTxs[i] == nil {
			gPackedTxs[i] = packedTx
			return C.int64_t(i)
		}
	}
	gPackedTxs = append(gPackedTxs, packedTx)
	return C.int64_t(len(gPackedTxs) - 1)
}

//export transaction_new_
func transaction_new_(expiration C.int64_t, refBlock *C.char, chainId *C.char) C.int64_t {
	tx := uuoskit.NewTransaction(int(expiration))
	tx.SetReferenceBlock(C.GoString(refBlock))

	packedTx := uuoskit.NewPackedTransaction(tx)
	packedTx.SetChainId(C.GoString(chainId))

	return addPackedTx(packedTx)
}

//export transaction_from_json_
func transaction_from_json_(tx *C.char, chainId *C.char) *C.char {
	_tx := C.GoString(tx)
	packedTx, err := uuoskit.NewPackedTransactionFromString(_tx)
	if err != nil {
		return renderError(err)
	}
	packedTx.SetChainId(C.GoString(chainId))
	i := addPackedTx(packedTx)
	return renderData(i)
}

//export transaction_free_
func transaction_free_(_index C.int64_t) *C.char {
	index := int(_index)
	if index < 0 || index >= len(gPackedTxs) {
		return renderError(errors.New("bad index"))
	}
	gPackedTxs[int(index)] = nil
	return renderData("ok")
}

//export transaction_set_chain_id_
func transaction_set_chain_id_(_index C.int64_t, chainId *C.char) *C.char {
	index := int(_index)
	if index < 0 || index >= len(gPackedTxs) {
		return renderError(fmt.Errorf("invalid index"))
	}
	err := gPackedTxs[int(index)].SetChainId(C.GoString(chainId))
	if err != nil {
		return renderError(err)
	}
	return renderData("ok")
}

//export transaction_add_action_
func transaction_add_action_(idx C.int64_t, account *C.char, name *C.char, data *C.char, permissions *C.char) *C.char {
	if err := validateIndex(idx); err != nil {
		return renderError(err)
	}

	_account := C.GoString(account)
	_name := C.GoString(name)
	_data := C.GoString(data)
	_permissions := C.GoString(permissions)

	var __data []byte
	__data, err := hex.DecodeString(_data)
	if err != nil {
		__data, err = uuoskit.GetABISerializer().PackActionArgs(_account, _name, []byte(_data))
		if err != nil {
			return renderError(err)
		}
	}

	perms := make(map[string]string)
	err = json.Unmarshal([]byte(_permissions), &perms)
	if err != nil {
		return renderError(err)
	}

	action := uuoskit.NewAction(uuoskit.NewName(_account), uuoskit.NewName(_name))
	action.SetData(__data)
	for k, v := range perms {
		action.AddPermission(uuoskit.NewName(k), uuoskit.NewName(v))
	}

	err = gPackedTxs[idx].AddAction(action)
	if err != nil {
		return renderError(err)
	}
	return renderData("ok")
}

//export transaction_sign_
func transaction_sign_(idx C.int64_t, pub *C.char) *C.char {
	if err := validateIndex(idx); err != nil {
		return renderError(err)
	}

	_pub := C.GoString(pub)
	sign, err := gPackedTxs[idx].Sign(_pub)
	if err != nil {
		return renderError(err)
	}
	return renderData(sign)
}

//export transaction_digest_
func transaction_digest_(idx C.int64_t, chainId *C.char) *C.char {
	if err := validateIndex(idx); err != nil {
		return renderError(err)
	}

	_chainId := C.GoString(chainId)
	digest, err := gPackedTxs[idx].Digest(_chainId)
	if err != nil {
		return renderError(err)
	}
	return renderData(digest)
}

//export transaction_sign_by_private_key_
func transaction_sign_by_private_key_(idx C.int64_t, priv *C.char) *C.char {
	//	func (t *PackedTransaction) SignByPrivateKey(privKey string, chainId string) (string, error) {
	if err := validateIndex(idx); err != nil {
		return renderError(err)
	}

	sign, err := gPackedTxs[idx].SignByPrivateKey(C.GoString(priv))
	if err != nil {
		return renderError(err)
	}
	return renderData(sign)
}

//export transaction_pack_
func transaction_pack_(idx C.int64_t, compress C.int) *C.char {
	if err := validateIndex(idx); err != nil {
		return renderError(err)
	}

	var result string
	if compress != 0 {
		result = gPackedTxs[idx].Pack(true)
	} else {
		result = gPackedTxs[idx].Pack(false)
	}
	return renderData(result)
}

//export transaction_marshal_
func transaction_marshal_(idx C.int64_t) *C.char {
	if err := validateIndex(idx); err != nil {
		return renderError(err)
	}

	result := gPackedTxs[idx].Marshal()
	return renderData(result)
}

//func (t *Transaction) Unpack(data []byte) (int, error) {
//export transaction_unpack_
func transaction_unpack_(data *C.char) *C.char {
	t := uuoskit.Transaction{}
	_data, err := hex.DecodeString(C.GoString(data))
	if err != nil {
		return renderError(err)
	}

	t.Unpack([]byte(_data))
	js, err := json.Marshal(t)
	if err != nil {
		return renderError(err)
	}
	return renderData(string(js))
}

//export abiserializer_set_contract_abi_
func abiserializer_set_contract_abi_(account *C.char, abi *C.char, length C.int) *C.char {
	_account := C.GoString(account)
	_abi := C.GoBytes(unsafe.Pointer(abi), length)
	err := uuoskit.GetABISerializer().SetContractABI(_account, _abi)
	if err != nil {
		return renderError(err)
	}
	return renderData("ok")
}

//export abiserializer_pack_action_args_
func abiserializer_pack_action_args_(contractName *C.char, actionName *C.char, args *C.char, args_len C.int) *C.char {
	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoBytes(unsafe.Pointer(args), args_len)
	result, err := uuoskit.GetABISerializer().PackActionArgs(_contractName, _actionName, _args)
	if err != nil {
		return renderError(err)
	}
	return renderData(hex.EncodeToString(result))
}

//export abiserializer_unpack_action_args_
func abiserializer_unpack_action_args_(contractName *C.char, actionName *C.char, args *C.char) *C.char {
	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoString(args)
	__args, err := hex.DecodeString(_args)
	if err != nil {
		return renderError(err)
	}
	result, err := uuoskit.GetABISerializer().UnpackActionArgs(_contractName, _actionName, __args)
	if err != nil {
		return renderError(err)
	}
	return renderData(string(result))
}

//export abiserializer_pack_abi_type_
func abiserializer_pack_abi_type_(contractName *C.char, actionName *C.char, args *C.char, args_len C.int) *C.char {
	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoBytes(unsafe.Pointer(args), args_len)
	result, err := uuoskit.GetABISerializer().PackAbiType(_contractName, _actionName, _args)
	if err != nil {
		return renderError(err)
	}
	return renderData(hex.EncodeToString(result))
}

//export abiserializer_unpack_abi_type_
func abiserializer_unpack_abi_type_(contractName *C.char, actionName *C.char, args *C.char) *C.char {
	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoString(args)
	__args, err := hex.DecodeString(_args)
	if err != nil {
		return renderError(err)
	}
	result, err := uuoskit.GetABISerializer().UnpackAbiType(_contractName, _actionName, __args)
	if err != nil {
		return renderError(err)
	}
	return renderData(string(result))
}

//export abiserializer_is_abi_cached_
func abiserializer_is_abi_cached_(contractName *C.char) C.int {
	_contractName := C.GoString(contractName)
	result := uuoskit.GetABISerializer().IsAbiCached(_contractName)
	if result {
		return 1
	} else {
		return 0
	}
}

//export s2n_
func s2n_(s *C.char) C.uint64_t {
	return C.uint64_t(uuoskit.S2N(C.GoString(s)))
}

//export n2s_
func n2s_(n C.uint64_t) *C.char {
	return C.CString(uuoskit.N2S(uint64(n)))
}

//symbol to uint64
//export sym2n_
func sym2n_(str_symbol *C.char, precision C.uint64_t) C.uint64_t {
	v := uuoskit.NewSymbol(C.GoString(str_symbol), int(uint64(precision))).Value
	return C.uint64_t(v)
}

//export abiserializer_pack_abi_
func abiserializer_pack_abi_(str_abi *C.char) *C.char {
	_str_abi := C.GoString(str_abi)
	result, err := uuoskit.GetABISerializer().PackABI(_str_abi)
	if err != nil {
		return renderError(err)
	}
	return renderData(hex.EncodeToString(result))
}

//export abiserializer_unpack_abi_
func abiserializer_unpack_abi_(abi *C.char, length C.int) *C.char {
	_abi := C.GoBytes(unsafe.Pointer(abi), length)
	result, err := uuoskit.GetABISerializer().UnpackABI(_abi)
	if err != nil {
		return renderError(err)
	}
	return renderData(result)
}

//export crypto_sign_digest_
func crypto_sign_digest_(digest *C.char, privateKey *C.char) *C.char {
	// log.Println(C.GoString(digest), C.GoString(privateKey))

	_privateKey, err := secp256k1.NewPrivateKeyFromBase58(C.GoString(privateKey))
	if err != nil {
		return renderError(err)
	}

	_digest, err := hex.DecodeString(C.GoString(digest))
	if err != nil {
		return renderError(err)
	}

	result, err := _privateKey.Sign(_digest)
	if err != nil {
		return renderError(err)
	}
	return renderData(result.String())
}

//export crypto_get_public_key_
func crypto_get_public_key_(privateKey *C.char, eosPub C.int) *C.char {
	_privateKey, err := secp256k1.NewPrivateKeyFromBase58(C.GoString(privateKey))
	if err != nil {
		return renderError(err)
	}

	pub := _privateKey.GetPublicKey()

	if eosPub == 0 {
		return renderData(pub.String())
	} else {
		return renderData(pub.StringEOS())
	}
}

//export crypto_recover_key_
func crypto_recover_key_(digest *C.char, signature *C.char, format C.int) *C.char {
	_digest, err := hex.DecodeString(C.GoString(digest))
	if err != nil {
		return renderError(err)
	}

	_signature, err := secp256k1.NewSignatureFromBase58(C.GoString(signature))
	if err != nil {
		return renderError(err)
	}

	pub, err := secp256k1.Recover(_digest, _signature)
	if err != nil {
		return renderError(err)
	}

	if format == 1 {
		return renderData(pub.StringEOS())
	} else {
		return renderData(pub.String())
	}
}

//export crypto_create_key_
func crypto_create_key_(oldPubKeyFormat C.bool) *C.char {
	ret := CreateKey(bool(oldPubKeyFormat))
	return renderData(ret)
}

func CreateKey(oldPubKeyFormat bool) map[string]string {
	priv := genPrivKey(rand.Reader)
	_priv := secp256k1.NewPrivateKey(priv)

	ret := make(map[string]string)
	ret["private"] = _priv.String()
	if oldPubKeyFormat {
		ret["public"] = _priv.GetPublicKey().StringEOS()
	} else {
		ret["public"] = _priv.GetPublicKey().String()
	}
	return ret
}

//https://github.com/tendermint/tendermint/blob/de2cffe7a44e99bf81a62c542d50ccbf28dc852a/crypto/secp256k1/secp256k1.go#L73
// genPrivKey generates a new secp256k1 private key using the provided reader.
func genPrivKey(rand io.Reader) []byte {
	var privKeyBytes [32]byte
	d := new(big.Int)

	//"github.com/btcsuite/btcd/btcec"
	//btcec.S256().N
	//log.Println(btcec.S256().N.Bytes())
	N := big.NewInt(0).SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254, 186, 174, 220, 230, 175, 72, 160, 59, 191, 210, 94, 140, 208, 54, 65, 65})

	for {
		privKeyBytes = [32]byte{}
		_, err := io.ReadFull(rand, privKeyBytes[:])
		if err != nil {
			panic(err)
		}

		d.SetBytes(privKeyBytes[:])
		// break if we found a valid point (i.e. > 0 and < N == curverOrder)
		// isValidFieldElement := 0 < d.Sign() && d.Cmp(btcec.S256().N) < 0
		isValidFieldElement := 0 < d.Sign() && d.Cmp(N) < 0
		if isValidFieldElement {
			break
		}
	}

	return privKeyBytes[:]
}

//export set_debug_flag_
func set_debug_flag_(debug C.bool) {
	uuoskit.SetDebug(bool(debug))
}

//export get_debug_flag_
func get_debug_flag_() C.bool {
	return C.bool(uuoskit.GetDebug())
}

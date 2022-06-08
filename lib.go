package main

//#include <stdint.h>
//#include <stddef.h>
//#include <stdbool.h>
// static char* get_p(char **pp, int i)
// {
//	    return pp[i];
// }
//typedef char *(*fn_malloc)(uint64_t size);
//static fn_malloc g_malloc = NULL;
//static void set_malloc_fn(fn_malloc fn) {
//	g_malloc = fn;
//}
//static char *cmalloc(uint64_t size) {
//	return 	g_malloc(size);
//}
import "C"

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	return CString(string(result))
}

func renderError(err error) *C.char {
	if _err, ok := err.(*traceable_errors.Error); ok {
		errMsg := _err.ErrorStack()
		ret := map[string]interface{}{"error": errMsg}
		result, _ := json.Marshal(ret)
		return CString(string(result))
	} else {
		pc, fn, line, _ := runtime.Caller(1)
		errMsg := fmt.Sprintf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
		ret := map[string]interface{}{"error": errMsg}
		result, _ := json.Marshal(ret)
		return CString(string(result))
	}
}

//export init_
func init_(malloc C.fn_malloc) {
	C.set_malloc_fn(malloc)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	secp256k1.Init()
}

func CString(s string) *C.char {
	p := C.cmalloc(C.uint64_t(len(s) + 1))
	pp := (*[1 << 30]byte)(unsafe.Pointer(p))
	copy(pp[:], s)
	pp[len(s)] = 0
	return (*C.char)(p)
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

//export wallet_remove_
func wallet_remove_(name *C.char, pubKey *C.char) C.bool {
	_name := C.GoString(name)
	_pubKey := C.GoString(pubKey)
	ret := uuoskit.GetWallet().Remove(_name, _pubKey)
	return C.bool(ret)
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

var gChainContexts []*uuoskit.ChainContext

//export new_chain_context_
func new_chain_context_() C.int64_t {
	if gChainContexts == nil {
		gChainContexts = make([]*uuoskit.ChainContext, 0, 64)
	}

	for i := 0; i < len(gChainContexts); i++ {
		if gChainContexts[i] == nil {
			gChainContexts[i] = uuoskit.NewChainContext()
			return C.int64_t(i)
		}
	}

	if len(gChainContexts) >= 64 {
		return C.int64_t(-1)
	}

	gChainContexts = append(gChainContexts, uuoskit.NewChainContext())
	return C.int64_t(len(gChainContexts) - 1)
}

//export chain_context_free_
func chain_context_free_(_index C.int64_t) *C.char {
	index := int(_index)
	if index < 0 || index >= len(gChainContexts) {
		return renderError(fmt.Errorf("bad chain index %d", index))
	}
	gChainContexts[int(index)] = nil
	return renderData("ok")
}

func getChainContext(index int) (*uuoskit.ChainContext, error) {
	if index < 0 || index >= len(gChainContexts) {
		return nil, fmt.Errorf("invalid chain index %d", index)
	}
	return gChainContexts[int(index)], nil
}

func validateIndex(packedTxs []*uuoskit.PackedTransaction, idx C.int64_t) error {
	if idx < 0 || idx >= C.int64_t(len(packedTxs)) {
		return fmt.Errorf("invalid transaction index %d", idx)
	}

	if packedTxs[idx] == nil {
		return fmt.Errorf("transaction at index %d is nil!", idx)
	}

	return nil
}

func addPackedTx(ctx *uuoskit.ChainContext, packedTx *uuoskit.PackedTransaction) C.int64_t {
	packedTxs := ctx.PackedTxs
	if packedTxs == nil {
		return -1
	}

	for i := 0; i < len(packedTxs); i++ {
		if packedTxs[i] == nil {
			packedTxs[i] = packedTx
			return C.int64_t(i)
		}
	}

	if len(packedTxs) >= 1024 {
		return C.int64_t(-1)
	}

	ctx.PackedTxs = append(ctx.PackedTxs, packedTx)
	return C.int64_t(len(ctx.PackedTxs) - 1)
}

//export transaction_new_
func transaction_new_(chainIndex C.int64_t, expiration C.int64_t, refBlock *C.char, chainId *C.char) C.int64_t {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return C.int64_t(-1)
	}
	tx := uuoskit.NewTransaction(int(expiration))
	tx.SetReferenceBlock(C.GoString(refBlock))

	packedTx := uuoskit.NewPackedTransaction(tx)
	packedTx.SetChainId(C.GoString(chainId))

	return addPackedTx(ctx, packedTx)
}

//export transaction_from_json_
func transaction_from_json_(chainIndex C.int64_t, tx *C.char, chainId *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_tx := C.GoString(tx)
	packedTx, err := uuoskit.NewPackedTransactionFromString(_tx)
	if err != nil {
		return renderError(err)
	}
	packedTx.SetChainId(C.GoString(chainId))
	i := addPackedTx(ctx, packedTx)
	return renderData(i)
}

//export transaction_free_
func transaction_free_(chainIndex C.int64_t, _index C.int64_t) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	index := int(_index)
	if index < 0 || index >= len(ctx.PackedTxs) {
		return renderError(fmt.Errorf("bad transaction index %d", index))
	}
	ctx.PackedTxs[int(index)] = nil
	return renderData("ok")
}

//export transaction_set_chain_id_
func transaction_set_chain_id_(chainIndex C.int64_t, _index C.int64_t, chainId *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	index := int(_index)
	if index < 0 || index >= len(ctx.PackedTxs) {
		return renderError(fmt.Errorf("invalid index"))
	}
	err = ctx.PackedTxs[int(index)].SetChainId(C.GoString(chainId))
	if err != nil {
		return renderError(err)
	}
	return renderData("ok")
}

//export transaction_add_action_
func transaction_add_action_(chainIndex C.int64_t, idx C.int64_t, account *C.char, name *C.char, data *C.char, permissions *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	if err := validateIndex(ctx.PackedTxs, idx); err != nil {
		return renderError(err)
	}

	_account := C.GoString(account)
	_name := C.GoString(name)
	_data := C.GoString(data)
	_permissions := C.GoString(permissions)

	var __data []byte
	__data, err = hex.DecodeString(_data)
	if err != nil {
		__data, err = ctx.ABISerializer.PackActionArgs(_account, _name, []byte(_data))
		if err != nil {
			return renderError(err)
		}
	}

	perms := make([]map[string]string, 0, 4)
	err = json.Unmarshal([]byte(_permissions), &perms)
	if err != nil {
		return renderError(err)
	}

	action := uuoskit.NewAction(uuoskit.NewName(_account), uuoskit.NewName(_name))
	action.SetData(__data)
	for _, item := range perms {
		for k, v := range item {
			action.AddPermission(uuoskit.NewName(k), uuoskit.NewName(v))
		}
	}

	err = ctx.PackedTxs[idx].AddAction(action)
	if err != nil {
		return renderError(err)
	}
	return renderData("ok")
}

//export transaction_sign_
func transaction_sign_(chainIndex C.int64_t, idx C.int64_t, pub *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	if err := validateIndex(ctx.PackedTxs, idx); err != nil {
		return renderError(err)
	}

	_pub := C.GoString(pub)
	sign, err := ctx.PackedTxs[idx].Sign(_pub)
	if err != nil {
		return renderError(err)
	}
	return renderData(sign)
}

//export transaction_digest_
func transaction_digest_(chainIndex C.int64_t, idx C.int64_t, chainId *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	if err := validateIndex(ctx.PackedTxs, idx); err != nil {
		return renderError(err)
	}

	_chainId := C.GoString(chainId)
	digest, err := ctx.PackedTxs[idx].Digest(_chainId)
	if err != nil {
		return renderError(err)
	}
	return renderData(digest)
}

//export transaction_sign_by_private_key_
func transaction_sign_by_private_key_(chainIndex C.int64_t, idx C.int64_t, priv *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	if err := validateIndex(ctx.PackedTxs, idx); err != nil {
		return renderError(err)
	}

	sign, err := ctx.PackedTxs[idx].SignByPrivateKey(C.GoString(priv))
	if err != nil {
		return renderError(err)
	}
	return renderData(sign)
}

//export transaction_pack_
func transaction_pack_(chainIndex C.int64_t, idx C.int64_t, compress C.int) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	if err := validateIndex(ctx.PackedTxs, idx); err != nil {
		return renderError(err)
	}

	var result string
	if compress != 0 {
		result = ctx.PackedTxs[idx].Pack(true)
	} else {
		result = ctx.PackedTxs[idx].Pack(false)
	}
	return renderData(result)
}

//export transaction_marshal_
func transaction_marshal_(chainIndex C.int64_t, idx C.int64_t) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	if err := validateIndex(ctx.PackedTxs, idx); err != nil {
		return renderError(err)
	}

	result := ctx.PackedTxs[idx].Marshal()
	return renderData(result)
}

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
func abiserializer_set_contract_abi_(chainIndex C.int64_t, account *C.char, abi *C.char, length C.int) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_account := C.GoString(account)
	_abi := C.GoBytes(unsafe.Pointer(abi), length)
	err = ctx.ABISerializer.SetContractABI(_account, _abi)
	if err != nil {
		return renderError(err)
	}
	return renderData("ok")
}

//export abiserializer_pack_action_args_
func abiserializer_pack_action_args_(chainIndex C.int64_t, contractName *C.char, actionName *C.char, args *C.char, args_len C.int) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoBytes(unsafe.Pointer(args), args_len)
	result, err := ctx.ABISerializer.PackActionArgs(_contractName, _actionName, _args)
	if err != nil {
		return renderError(err)
	}
	return renderData(hex.EncodeToString(result))
}

//export abiserializer_unpack_action_args_
func abiserializer_unpack_action_args_(chainIndex C.int64_t, contractName *C.char, actionName *C.char, args *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoString(args)
	__args, err := hex.DecodeString(_args)
	if err != nil {
		return renderError(err)
	}
	result, err := ctx.ABISerializer.UnpackActionArgs(_contractName, _actionName, __args)
	if err != nil {
		return renderError(err)
	}
	return renderData(string(result))
}

//export abiserializer_pack_abi_type_
func abiserializer_pack_abi_type_(chainIndex C.int64_t, contractName *C.char, actionName *C.char, args *C.char, args_len C.int) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoBytes(unsafe.Pointer(args), args_len)
	result, err := ctx.ABISerializer.PackAbiType(_contractName, _actionName, _args)
	if err != nil {
		return renderError(err)
	}
	return renderData(hex.EncodeToString(result))
}

//export abiserializer_unpack_abi_type_
func abiserializer_unpack_abi_type_(chainIndex C.int64_t, contractName *C.char, actionName *C.char, args *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_contractName := C.GoString(contractName)
	_actionName := C.GoString(actionName)
	_args := C.GoString(args)
	__args, err := hex.DecodeString(_args)
	if err != nil {
		return renderError(err)
	}
	result, err := ctx.ABISerializer.UnpackAbiType(_contractName, _actionName, __args)
	if err != nil {
		return renderError(err)
	}
	return renderData(string(result))
}

//export abiserializer_is_abi_cached_
func abiserializer_is_abi_cached_(chainIndex C.int64_t, contractName *C.char) C.int {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return 0
	}

	_contractName := C.GoString(contractName)
	result := ctx.ABISerializer.IsAbiCached(_contractName)
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
	return CString(uuoskit.N2S(uint64(n)))
}

//symbol to uint64
//export sym2n_
func sym2n_(str_symbol *C.char, precision C.uint64_t) C.uint64_t {
	v := uuoskit.NewSymbol(C.GoString(str_symbol), int(uint64(precision))).Value
	return C.uint64_t(v)
}

//export abiserializer_pack_abi_
func abiserializer_pack_abi_(chainIndex C.int64_t, str_abi *C.char) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_str_abi := C.GoString(str_abi)
	result, err := ctx.ABISerializer.PackABI(_str_abi)
	if err != nil {
		return renderError(err)
	}
	return renderData(hex.EncodeToString(result))
}

//export abiserializer_unpack_abi_
func abiserializer_unpack_abi_(chainIndex C.int64_t, abi *C.char, length C.int) *C.char {
	ctx, err := getChainContext(int(chainIndex))
	if err != nil {
		return renderError(err)
	}

	_abi := C.GoBytes(unsafe.Pointer(abi), length)
	result, err := ctx.ABISerializer.UnpackABI(_abi)
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

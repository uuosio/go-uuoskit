package main

// #include <stdint.h>
// static char* get_p(char **pp, int i)
// {
//	    return pp[i];
// }
import "C"

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
)

func renderData(data interface{}) *C.char {
	ret := map[string]interface{}{"data": data}
	result, _ := json.Marshal(ret)
	return C.CString(string(result))
}

func renderError(err error) *C.char {
	pc, fn, line, _ := runtime.Caller(1)
	errMsg := fmt.Sprintf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
	ret := map[string]interface{}{"error": errMsg}
	result, _ := json.Marshal(ret)
	return C.CString(string(result))
}

//export init_
func init_() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
	err := GetWallet().Import(_name, _priv)
	if err != nil {
		return renderError(err)
	}
	log.Println("import", _name, _priv)
	return renderData("ok")
}

var gPackedTxs []*PackedTransaction

//export transaction_new_
func transaction_new_(expiration C.int64_t, refBlock *C.char, chainId *C.char) C.int64_t {
	tx := NewTransaction(int(expiration))
	tx.SetReferenceBlock(C.GoString(refBlock))

	packedTx := NewPackedtransaction(tx)
	packedTx.SetChainId(C.GoString(chainId))
	if gPackedTxs == nil {
		//element at 0 not used
		gPackedTxs = make([]*PackedTransaction, 0, 10)
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

//export transaction_free_
func transaction_free_(_index C.int64_t) {
	index := int(_index)
	if index < 0 || index >= len(gPackedTxs) {
		return
	}
	gPackedTxs[int(index)] = nil
	return
}

//export transaction_add_action_
func transaction_add_action_(idx C.int64_t, account *C.char, name *C.char, data *C.char, permissions *C.char) *C.char {
	if idx < 0 || idx >= C.int64_t(len(gPackedTxs)) {
		return renderError(fmt.Errorf("invalid idx"))
	}
	if gPackedTxs[idx] == nil {
		return renderError(fmt.Errorf("invalid idx"))
	}
	_account := C.GoString(account)
	_name := C.GoString(name)
	_data := C.GoString(data)
	_permissions := C.GoString(permissions)

	var __data []byte
	__data, err := hex.DecodeString(_data)
	if err != nil {
		__data, err = GetABISerializer().PackActionArgs(_account, _name, _data)
		if err != nil {
			return renderError(err)
		}
	}

	perms := make(map[string]string)
	err = json.Unmarshal([]byte(_permissions), &perms)
	if err != nil {
		return renderError(err)
	}

	action := NewAction(NewName(_account), NewName(_name))
	action.SetData(__data)
	for k, v := range perms {
		action.AddPermission(NewName(k), NewName(v))
	}
	gPackedTxs[idx].tx.AddAction(action)
	return renderData("ok")
}

//export transaction_sign_
func transaction_sign_(idx C.int64_t, pub *C.char) *C.char {
	if idx < 0 || idx >= C.int64_t(len(gPackedTxs)) {
		return renderError(fmt.Errorf("invalid idx"))
	}

	if gPackedTxs[idx] == nil {
		return renderError(fmt.Errorf("invalid idx"))
	}

	_pub := C.GoString(pub)
	err := gPackedTxs[idx].Sign(_pub)
	if err != nil {
		return renderError(err)
	}
	return renderData("ok")
}

// func NewPackedtransaction(tx *Transaction) *PackedTransaction {

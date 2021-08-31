package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"time"

	"fmt"
	"testing"

	"github.com/iancoleman/orderedmap"
	"github.com/learnforpractice/go-secp256k1/secp256k1"
)

func TestOrderedMap(t *testing.T) {
	t.Log("hello,world")
	o := orderedmap.New()
	json.Unmarshal([]byte(`{"hello":123, "a": "b"}`), &o)
	fmt.Println(o)
	s, _ := json.Marshal(o)
	fmt.Println(string(s))
	for k, v := range o.Keys() {
		fmt.Println("+++++++k, v:", k, v)
	}
}

func TestCrypto(t *testing.T) {
	secp256k1.Init()
	defer secp256k1.Destroy()

	// digest := make([]byte, 32)
	//	seckey := make([]byte, 32)
	digest, err := hex.DecodeString("2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	if err != nil {
		panic(err)
	}

	seckey, err := secp256k1.NewPrivateKeyFromHex("99870ba61ad4bfae18a1c4cea5a6b48882b95633421b108497a2b53dc779a639")
	if err != nil {
		panic(err)
	}

	start := time.Now()

	signature, err := secp256k1.Sign(digest, seckey)
	if err != nil {
		panic(err)
	}
	log.Println("++++++signature:", signature.String())

	{
		pubKey, _ := secp256k1.Recover(digest, signature)
		log.Println("++++++pubKeyBase58:", pubKey.String())
	}

	{
		pubkey, _ := secp256k1.GetPublicKey(seckey)
		log.Println("++++++pub key:", pubkey.String())
		duration := time.Since(start)
		log.Println("++++++duration:", duration)
	}
}

func TestPackTransaction(t *testing.T) {
	tx := `{"expiration":"2021-08-31T05:59:39","ref_block_num":56745,"ref_block_prefix":3729394962,"max_net_usage_words":0,"max_cpu_usage_ms":0,"delay_sec":0,"context_free_actions":[],"actions":[{"account":"eosio.token","name":"transfer","authorization":[{"actor":"helloworld11","permission":"active"}],"data":"10428a97721aa36a0000000000000e3d102700000000000004454f53000000001a7472616e736665722066726f6d20616c69636520746f20626f62"}],"transaction_extensions":[],"signatures":["SIG_K1_KbSF8BCNVA95KzR1qLmdn4VnxRoLVFQ1fZ8VV5gVdW1hLfGBdcwEc93hF7FBkWZip1tq2Ps27UZxceaR3hYwAjKL7j59q8"],"context_free_data":[]}`
	o := orderedmap.New()
	json.Unmarshal([]byte(tx), &o)
	s, _ := json.Marshal(o)
	fmt.Println(string(s))
	for _, k := range o.Keys() {
		v, exist := o.Get(k)
		fmt.Println("+++++++k, v:", exist, k, v)
	}
}

func TestAbi(t *testing.T) {
	serializer := NewABISerializer()
	fieldType := serializer.GetType("transaction", "expiration")
	t.Log("+++++++fieldType:", fieldType)

	fieldType = serializer.GetType("transaction", "transaction_extensions")
	t.Log("+++++++fieldType:", fieldType)

	strAbi, err := os.ReadFile("data/eosio.token.abi")
	if err != nil {
		panic(err)
	}
	serializer.AddContractAbi("hello", strAbi)
	actionStruct := serializer.FindActionType("hello", "transfer")
	t.Log("+++++++actionType:", actionStruct)
}

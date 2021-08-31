package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
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

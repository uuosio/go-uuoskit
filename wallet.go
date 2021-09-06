package main

import (
	"errors"

	"github.com/uuosio/go-secp256k1/secp256k1"
)

type Wallet struct {
	keys map[string]*secp256k1.PrivateKey
}

var gWallet *Wallet

func GetWallet() *Wallet {
	if gWallet == nil {
		gWallet = &Wallet{}
		gWallet.keys = make(map[string]*secp256k1.PrivateKey)
	}
	return gWallet
}

func (w *Wallet) Import(name string, strPriv string) error {
	priv, err := secp256k1.NewPrivateKeyFromBase58(strPriv)
	if err != nil {
		return err
	}

	pub := priv.GetPublicKey()
	w.keys[pub.String()] = priv
	return nil
}

//GetPublicKeys
func (w *Wallet) GetPublicKeys() []string {
	keys := make([]string, 0, len(w.keys))
	for k := range w.keys {
		keys = append(keys, k)
	}
	return keys
}

func (w *Wallet) GetPrivateKey(pubKey string) (*secp256k1.PrivateKey, error) {
	priv, ok := w.keys[pubKey]
	if !ok {
		return nil, errors.New("not found")
	}
	return priv, nil
}

func (w *Wallet) Sign(digest []byte, pubKey string) (*secp256k1.Signature, error) {
	priv, ok := w.keys[pubKey]
	if !ok {
		return nil, errors.New("not found")
	}
	sig, err := priv.Sign(digest)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

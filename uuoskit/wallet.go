package uuoskit

import (
	secp256k1 "github.com/uuosio/go-secp256k1"
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
		return newError(err)
	}

	pub := priv.GetPublicKey()
	w.keys[pub.StringEOS()] = priv
	return nil
}

func (w *Wallet) Remove(name string, pubKey string) bool {
	_pubKey, err := secp256k1.NewPublicKeyFromBase58(pubKey)
	if err != nil {
		return false
	}

	pubKey = _pubKey.StringEOS()
	if priv, ok := w.keys[pubKey]; ok {
		for i := 0; i < len(w.keys); i++ {
			priv.Data[i] = 0
		}
		delete(w.keys, pubKey)
		return true
	}
	return false
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
		return nil, newErrorf("not found")
	}
	return priv, nil
}

func (w *Wallet) Sign(digest []byte, pubKey string) (*secp256k1.Signature, error) {
	pub, err := secp256k1.NewPublicKeyFromBase58(pubKey)
	if err != nil {
		return nil, newError(err)
	}

	priv, ok := w.keys[pub.StringEOS()]
	if !ok {
		return nil, newErrorf("not found")
	}
	sig, err := priv.Sign(digest)
	if err != nil {
		return nil, newError(err)
	}
	return sig, nil
}

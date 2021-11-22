package uuoskit

import (
	"testing"
	"time"
)

func TestGetRequiredKeys(t *testing.T) {
	priv := "5JRYimgLBrRLCBAcjHUWCYRv3asNedTYYzVgmiU4q2ZVxMBiJXL"
	GetWallet().Import("test", priv)

	rpc := NewRpc("https://testnode.uuos.network:8443")

	chainInfo, err := rpc.GetInfo()
	if err != nil {
		panic(err)
	}
	expiration := int(time.Now().Unix()) + 60
	tx := NewTransaction(expiration)
	tx.SetReferenceBlock(chainInfo.LastIrreversibleBlockID)

	action := NewAction(NewName("eosio.token"),
		NewName("transfer"),
		[]PermissionLevel{{NewName("helloworld11"), NewName("active")}},
		NewName("helloworld11"),
		NewName("eosio.token"),
		NewAsset(1000, NewSymbol("EOS", 4)),
		"transfer from alice")
	action.AddPermission(NewName("helloworld11"), NewName("active"))
	tx.AddAction(action)

	args := GetRequiredKeysArgs{tx, GetWallet().GetPublicKeys()}
	ret, err := rpc.GetRequiredKeys(&args)
	if err != nil {
		panic(err)
	}

	t.Log(ret.RequiredKeys)
}

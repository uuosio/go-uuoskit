package uuoskit

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"time"

	"github.com/iancoleman/orderedmap"
)

type ChainApi struct {
	rpc *Rpc
}

func NewChainApi(rpcUrl string) *ChainApi {
	rpc := NewRpc(rpcUrl)
	chainApi := &ChainApi{rpc: rpc}
	return chainApi
}

func (api *ChainApi) GetTableRows(
	json bool,
	code string,
	scope string,
	table string,
	lowerbound string,
	upperbound string,
	limit int,
	keyType string,
	indexPosition int,
	reverse bool,
	showPayer bool,
) (*orderedmap.OrderedMap, error) {
	args := GetTableRowsArgs{
		Json:          json,
		Code:          code,
		Scope:         scope,
		Table:         table,
		LowerBound:    lowerbound,
		UpperBound:    upperbound,
		Limit:         limit,
		KeyType:       keyType,
		IndexPosition: indexPosition,
		Reverse:       reverse,
		ShowPayer:     showPayer,
	}
	return api.rpc.GetTableRows(&args)
}

func (api *ChainApi) DeployContract(account, codeFile string, abiFile string) error {
	code, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return err
	}

	abi, err := ioutil.ReadFile(abiFile)
	if err != nil {
		return err
	}

	binABI, err := GetABISerializer().PackABI(string(abi))
	if err != nil {
		return err
	}

	chainInfo, err := api.rpc.GetInfo()
	if err != nil {
		return err
	}

	expiration := int(time.Now().Unix()) + 60
	tx := NewTransaction(expiration)
	tx.SetReferenceBlock(chainInfo.LastIrreversibleBlockID)

	action := NewAction(
		NewName("eosio"),
		NewName("setcode"),
		[]PermissionLevel{{NewName(account), NewName("active")}},
		NewName(account),
		uint8(0), //vm_type
		uint8(0), //vm_version
		code,     //code
	)
	tx.AddAction(action)

	action = NewAction(
		NewName("eosio"),
		NewName("setabi"),
		[]PermissionLevel{{NewName(account), NewName("active")}},
		NewName(account), //account
		binABI,           //code
	)
	tx.AddAction(action)

	chainId := chainInfo.ChainID
	packedTx := NewPackedTransaction(tx)
	packedTx.SetChainId(chainId)

	args := GetRequiredKeysArgs{
		Transaction:   tx,
		AvailableKeys: GetWallet().GetPublicKeys(),
	}
	r, err := api.rpc.GetRequiredKeys(args)
	if err != nil {
		return err
	}

	for i := range r.RequiredKeys {
		pub := r.RequiredKeys[i]
		_, err = packedTx.Sign(pub)
		if err != nil {
			return err
		}
	}

	r2, err := api.rpc.PushTransaction(packedTx)
	if err != nil {
		return err
	}

	if _, ok := r2.Get("error"); ok {
		if msg, ok := DeepGet(r2, "error", "details", 0, "message"); ok {
			log.Println(msg)
			if msg == "contract is already running this version of code" {
				return nil
			}
		}
		r, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			panic(err)
		}
		return errors.New(string(r))
	}
	return nil
}

func (api *ChainApi) PushAction(account, action, args string, actor, permission string) (string, error) {
	chainInfo, err := api.rpc.GetInfo()
	if err != nil {
		return "", err
	}

	expiration := int(time.Now().Unix()) + 60
	tx := NewTransaction(expiration)
	tx.SetReferenceBlock(chainInfo.LastIrreversibleBlockID)

	result, err := GetABISerializer().PackActionArgs(account, action, []byte(args))
	if err != nil {
		return "", err
	}
	a := NewAction(
		NewName(account),
		NewName(action),
	)
	a.Data = result
	a.AddPermission(NewName(actor), NewName(permission))
	tx.AddAction(a)

	chainId := chainInfo.ChainID

	//FIXME: hardcode public key
	packedTx := NewPackedTransaction(tx)
	packedTx.SetChainId(chainId)

	rpcArgs := GetRequiredKeysArgs{
		Transaction:   tx,
		AvailableKeys: GetWallet().GetPublicKeys(),
	}
	r, err := api.rpc.GetRequiredKeys(rpcArgs)
	if err != nil {
		return "", err
	}

	for i := range r.RequiredKeys {
		pub := r.RequiredKeys[i]
		_, err = packedTx.Sign(pub)
		if err != nil {
			return "", err
		}
	}

	r2, err := api.rpc.PushTransaction(packedTx)
	if err != nil {
		return "", err
	}

	r3, err := json.MarshalIndent(r2, "", "  ")
	if err != nil {
		panic(err)
	}

	if _, ok := r2.Get("error"); ok {
		return "", errors.New(string(r3))
	}
	return string(r3), nil
}

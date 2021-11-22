package uuoskit

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

type ChainApi struct {
	rpc *Rpc
}

func NewChainApi(rpcUrl string) *ChainApi {
	rpc := NewRpc(rpcUrl)
	chainApi := &ChainApi{rpc: rpc}
	return chainApi
}

func (api *ChainApi) GetAccount(name string) (JsonValue, error) {
	return api.rpc.GetAccount(&GetAccountArgs{AccountName: name})
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
) (JsonValue, error) {
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
		return newError(err)
	}

	abi, err := ioutil.ReadFile(abiFile)
	if err != nil {
		return newError(err)
	}

	binABI, err := GetABISerializer().PackABI(string(abi))
	if err != nil {
		return newError(err)
	}

	chainInfo, err := api.rpc.GetInfo()
	if err != nil {
		return newError(err)
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
	r, err := api.rpc.GetRequiredKeys(&args)
	if err != nil {
		return newError(err)
	}

	for i := range r.RequiredKeys {
		pub := r.RequiredKeys[i]
		_, err = packedTx.Sign(pub)
		if err != nil {
			return newError(err)
		}
	}

	r2, err := api.rpc.PushTransaction(packedTx)
	if err != nil {
		return newError(err)
	}

	if _, err := r2.Get("error"); err == nil {
		if msg, err := r2.Get("error", "details", 0, "message"); err == nil {
			log.Println(msg)
			if msg == "contract is already running this version of code" {
				return nil
			}
		}
		r, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			panic(err)
		}
		return newErrorf(string(r))
	}
	return nil
}

func (api *ChainApi) getRequiredKeys(actions []Action) ([]string, error) {
	args := GetRequiredKeysArgs{
		Transaction:   NewTransaction(0),
		AvailableKeys: GetWallet().GetPublicKeys(),
	}
	for i := range actions {
		a := actions[i]
		a.Data = []byte{}
		args.Transaction.AddAction(&a)
	}
	r, err := api.rpc.GetRequiredKeys(&args)
	if err != nil {
		return nil, newError(err)
	}
	return r.RequiredKeys, nil
}

func (api *ChainApi) PushActionWithArgs(account, action, args string, actor, permission string) (JsonValue, error) {
	result, err := GetABISerializer().PackActionArgs(account, action, []byte(args))
	if err != nil {
		return JsonValue{}, err
	}
	a := NewAction(
		NewName(account),
		NewName(action),
	)
	a.Data = result
	a.AddPermission(NewName(actor), NewName(permission))
	return api.PushAction(a)
}

func (api *ChainApi) PushAction(action *Action) (JsonValue, error) {
	return api.PushActions([]*Action{action})
}

func (api *ChainApi) PushActions(actions []*Action) (JsonValue, error) {
	chainInfo, err := api.rpc.GetInfo()
	if err != nil {
		return JsonValue{}, err
	}

	expiration := int(time.Now().Unix()) + 60
	tx := NewTransaction(expiration)
	tx.SetReferenceBlock(chainInfo.LastIrreversibleBlockID)
	for i := range actions {
		a := actions[i]
		tx.AddAction(a)
	}

	chainId := chainInfo.ChainID

	//FIXME: hardcode public key
	packedTx := NewPackedTransaction(tx)
	packedTx.SetChainId(chainId)

	pubKeys, err := api.getRequiredKeys(tx.Actions)
	if err != nil {
		return JsonValue{}, err
	}

	for i := range pubKeys {
		pub := pubKeys[i]
		_, err = packedTx.Sign(pub)
		if err != nil {
			return JsonValue{}, err
		}
	}

	r2, err := api.rpc.PushTransaction(packedTx)
	if err != nil {
		return JsonValue{}, err
	}
	if _, err := r2.Get("error"); err == nil {
		msg, err := r2.Get("error", "details", 0, "message")
		if err == nil {
			return r2, newErrorf(msg.(string))
		}
		log.Println(err)
		return r2, newErrorf("push_transaction error")
	}
	return r2, nil
}

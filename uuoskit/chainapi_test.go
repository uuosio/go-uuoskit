package uuoskit

import "testing"

func TestChainApi(t *testing.T) {
	GetWallet().Import("test", "5JRYimgLBrRLCBAcjHUWCYRv3asNedTYYzVgmiU4q2ZVxMBiJXL")
	api := NewChainApi("https://testnode.uuos.network:8443")
	a, _ := api.GetAccount("eosio")
	v, _ := a.Get("created")
	t.Logf("++%v\n", v)
	v, _ = a.Get("last_code_update")
	t.Logf("++%v\n", v)
}

func TestPushAction(t *testing.T) {
	GetWallet().Import("test", "5JRYimgLBrRLCBAcjHUWCYRv3asNedTYYzVgmiU4q2ZVxMBiJXL")
	api := NewChainApi("https://testnode.uuos.network:8443")
	action := NewAction(NewName("eosio.token"),
		NewName("transfer"),
		[]PermissionLevel{{NewName("helloworld11"), NewName("active")}},
		NewName("helloworld11"),
		NewName("eosio.token"),
		NewAsset(1000, NewSymbol("EOS", 4)),
		"transfer from alice")
	r, err := api.PushAction(action)
	if err != nil {
		t.Error(err)
	}

	if false {
		t.Log(r)
	}

	strAction := `
	{
		"from": "helloworld11",
		"to": "eosio.token",
		"quantity": "0.1000 EOS",
		"memo": "transfer from alice"
	}
	`
	r, err = api.PushActionWithArgs("eosio.token", "transfer", strAction, "helloworld11", "active")
	if err != nil {
		t.Error(err)
	}

	if false {
		t.Log(r)
	}
}

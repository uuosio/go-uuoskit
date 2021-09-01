package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/iancoleman/orderedmap"
)

type RpcError struct {
	err string
}

func (r *RpcError) Error() string {
	return r.err
}

func NewRpcError(err string) *RpcError {
	return &RpcError{err}
}

type Rpc struct {
	client *http.Client
	url    string
}

func NewRpc(url string) *Rpc {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	rpc := &Rpc{}
	rpc.url = url
	rpc.client = &http.Client{Transport: tr}
	return rpc
}

func (r *Rpc) GetInfo() (*ChainInfo, error) {
	var info ChainInfo
	err := r.Call("chain", "get_info", "", &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (r *Rpc) PushTransaction(packedTx interface{}) (*orderedmap.OrderedMap, error) {
	result := orderedmap.New()
	err := r.Call("chain", "push_transaction", packedTx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Rpc) Call(api string, endpoint string, params interface{}, result interface{}) error {
	var _params []byte
	reqUrl := fmt.Sprintf("%s/v1/%s/%s", r.url, api, endpoint)

	switch v := params.(type) {
	case string:
		_params = []byte(v)
	case []byte:
		_params = v
	default:
		var err error
		_params, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}

	if len(_params) == 0 {
		log.Println("++++reqUrl:", reqUrl)
		resp, err := r.client.Get(reqUrl)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, result)
		if err != nil {
			if err != nil {
				return NewRpcError(string(body))
			}
		}
		return nil
	}
	buf := bytes.NewBuffer([]byte(_params))
	resp, err := r.client.Post(reqUrl, "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// log.Println(string(body))

	err = json.Unmarshal(body, result)
	if err != nil {
		return NewRpcError(string(body))
	}
	return nil
}

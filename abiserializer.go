package main

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"
)

type ABITable struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	IndexType string   `json:"index_type"`
	KeyNames  []string `json:"key_names"`
	KeyTypes  []string `json:"key_types"`
}

type ABIAction struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	RicardianContract string `json:"ricardian_contract"`
}

type ABIStructField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ABIStruct struct {
	Name   string           `json:"name"`
	Base   string           `json:"base"`
	Fields []ABIStructField `json:"fields"`
}

type ABI struct {
	Version          string        `json:"version"`
	Structs          []ABIStruct   `json:"structs"`
	Types            []string      `json:"types"`
	Actions          []ABIAction   `json:"actions"`
	Tables           []ABITable    `json:"tables"`
	RicardianClauses []interface{} `json:"ricardian_clauses"`
	Variants         []interface{} `json:"variants"`
	AbiExtensions    []interface{} `json:"abi_extensions"`
	ErrorMessages    []interface{} `json:"error_messages"`
}

type ABISerializer struct {
	abiMap         map[string]*ABIStruct
	baseTypeMap    map[string]bool
	contractAbiMap map[string]*ABI
}

func NewABISerializer() *ABISerializer {
	serializer := &ABISerializer{}

	abi := ABI{}
	err := json.Unmarshal([]byte(baseABI), &abi)
	if err != nil {
		panic(err)
	}

	serializer.abiMap = make(map[string]*ABIStruct)
	serializer.baseTypeMap = make(map[string]bool)
	serializer.contractAbiMap = make(map[string]*ABI)

	for _, typeName := range baseTypes {
		serializer.baseTypeMap[typeName] = true
	}

	for i := range abi.Structs {
		s := &abi.Structs[i]
		serializer.abiMap[s.Name] = s
	}

	return serializer
}

func (t *ABISerializer) GetType(structName string, fieldName string) string {
	s := t.abiMap[structName]
	for _, f := range s.Fields {
		if f.Name == fieldName {
			return f.Type
		}
	}

	if s.Base != "" {
		return t.GetType(s.Base, fieldName)
	}
	return ""
}

func (t *ABISerializer) AddContractAbi(structName string, abi []byte) {
	abiObj := ABI{}
	err := json.Unmarshal(abi, &abiObj)
	if err != nil {
		panic(err)
	}

	t.contractAbiMap[structName] = &abiObj
}

func (t *ABISerializer) Pack(m orderedmap.OrderedMap) []byte {
	return nil
}

func (t *ABISerializer) Unpack(b []byte, structName string) string {
	return ""
}

func (t *ABISerializer) FindActionType(contractName string, actionName string) *ABIStruct {
	abi := t.contractAbiMap[contractName]
	for i := range abi.Actions {
		action := &abi.Actions[i]
		if action.Name == actionName {
			for j := range abi.Structs {
				s := &abi.Structs[j]
				if s.Name == action.Type {
					return s
				}
			}
		}
	}
	return nil
}

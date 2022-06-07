package uuoskit

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"
)

type ABISerializer struct {
	contractAbiMap map[string]*ABI
	contractName   string
}

func NewABISerializer() *ABISerializer {
	serializer := &ABISerializer{}
	serializer.contractAbiMap = make(map[string]*ABI)
	serializer.SetContractABI("eosio.token", []byte(eosioTokenAbi))
	return serializer
}

func (t *ABISerializer) SetContractABI(contractName string, abi []byte) error {
	if len(abi) == 0 {
		if _, ok := t.contractAbiMap[contractName]; ok {
			delete(t.contractAbiMap, contractName)
		}
		return nil
	}
	abiObj := &ABI{}
	err := json.Unmarshal(abi, abiObj)
	if err != nil {
		return newError(err)
	}

	t.contractAbiMap[contractName] = abiObj
	return nil
}

func (t *ABISerializer) IsAbiCached(contractName string) bool {
	_, ok := t.contractAbiMap[contractName]
	return ok
}

func (t *ABISerializer) PackActionArgs(contractName, actionName string, args []byte) ([]byte, error) {
	if abi, ok := t.contractAbiMap[contractName]; ok {
		actionTypeName := abi.GetActionStructType(actionName)
		return abi.PackAbiType(actionTypeName, args)
	} else {
		return nil, newErrorf("abi struct not found in %s::%s", contractName, actionName)
	}
}

func (t *ABISerializer) PackAbiStructByName(contractName, structName string, args string) ([]byte, error) {
	if abi, ok := t.contractAbiMap[contractName]; ok {
		return abi.PackAbiStructByName(structName, args)
	} else {
		return nil, newErrorf("abi struct not found in %s::%s", contractName, structName)
	}
}

func (t *ABISerializer) UnpackActionArgs(contractName string, actionName string, packedValue []byte) ([]byte, error) {
	abi, ok := t.contractAbiMap[contractName]
	if !ok {
		return nil, newErrorf("contract not found %s", contractName)
	}

	actionType := abi.GetActionStructType(actionName)
	if actionType == "" {
		return nil, newErrorf("unknown action %s::%s", contractName, actionName)
	}
	dec := NewDecoder(packedValue)
	result := orderedmap.New()
	err := abi.UnpackAbiStruct(dec, actionType, result)
	if err != nil {
		return nil, newError(err)
	}
	bs, err := json.Marshal(result)
	if err != nil {
		return nil, newError(err)
	}
	return bs, nil
}

func (t *ABISerializer) PackAbiType(contractName, abiType string, args []byte) ([]byte, error) {
	abi, ok := t.contractAbiMap[contractName]
	if !ok {
		return nil, newErrorf("contract not found %s", contractName)
	}
	return abi.PackAbiType(abiType, args)
}

func (t *ABISerializer) UnpackAbiType(contractName, abiName string, packedValue []byte) ([]byte, error) {
	abi, ok := t.contractAbiMap[contractName]
	if !ok {
		return nil, newErrorf("contract not found %s", contractName)
	}
	return abi.UnpackAbiType(abiName, packedValue)
}

func (t *ABISerializer) PackABI(strABI string) ([]byte, error) {
	abi := &ABI{}
	err := json.Unmarshal([]byte(strABI), abi)
	if err != nil {
		return nil, newError(err)
	}

	enc := NewEncoder(len(strABI) * 3)
	enc.PackString(abi.Version)
	enc.PackVarUint32(uint32(len(abi.Types)))
	for i := range abi.Types {
		t := &abi.Types[i]
		enc.PackString(t.NewTypeName)
		enc.PackString(t.Type)
	}

	enc.PackVarUint32(uint32(len(abi.Structs)))
	for i := range abi.Structs {
		s := &abi.Structs[i]
		enc.PackString(s.Name)
		enc.PackString(s.Base)
		enc.PackVarUint32(uint32(len(s.Fields)))
		for j := range s.Fields {
			f := &s.Fields[j]
			enc.PackString(f.Name)
			enc.PackString(f.Type)
		}
	}

	enc.PackVarUint32(uint32(len(abi.Actions)))
	for i := range abi.Actions {
		a := &abi.Actions[i]
		name := NewName(a.Name)
		enc.PackName(name)
		enc.PackString(a.Type)
		enc.PackString(a.RicardianContract)
	}

	enc.PackVarUint32(uint32(len(abi.Tables)))
	for i := range abi.Tables {
		t := &abi.Tables[i]
		name := NewName(t.Name)
		enc.PackName(name)
		enc.PackString(t.IndexType)
		enc.PackVarUint32(uint32(len(t.KeyTypes)))
		for j := range t.KeyNames {
			enc.PackString(t.KeyNames[j])
		}
		enc.PackVarUint32(uint32(len(t.KeyTypes)))
		for j := range t.KeyTypes {
			enc.PackString(t.KeyTypes[j])
		}
		enc.PackString(t.Type)
	}

	enc.PackVarUint32(uint32(len(abi.RicardianClauses)))
	for i := range abi.RicardianClauses {
		a := &abi.RicardianClauses[i]
		enc.PackString(a.Id)
		enc.PackString(a.Body)
	}

	enc.PackVarUint32(uint32(len(abi.ErrorMessages)))
	for i := range abi.ErrorMessages {
		a := &abi.ErrorMessages[i]
		enc.PackUint64(a.ErrorCode)
		enc.PackString(a.ErrorMsg)
	}

	enc.PackVarUint32(uint32(len(abi.AbiExtensions)))
	for i := range abi.AbiExtensions {
		a := &abi.AbiExtensions[i]
		enc.PackUint16(a.Type)
		enc.PackBytes(a.Extension)
	}

	enc.PackVarUint32(uint32(len(abi.Variants)))
	for i := range abi.Variants {
		a := &abi.Variants[i]
		enc.PackString(a.Name)
		enc.PackVarUint32(uint32(len(a.Types)))
		for i := range a.Types {
			tp := a.Types[i]
			enc.PackString(tp)
		}
	}

	return enc.Bytes(), nil
}

func (t *ABISerializer) UnpackABI(rawAbi []byte) (string, error) {
	dec := NewDecoder(rawAbi)
	abi := &ABI{}
	abi.Types = []ABIType{}
	abi.Structs = []ABIStruct{}
	abi.Actions = []ABIAction{}
	abi.Tables = []ABITable{}
	abi.RicardianClauses = []ClausePair{}
	abi.ErrorMessages = []ErrorMessage{}
	abi.AbiExtensions = []AbiExtension{}
	abi.Variants = []VariantDef{}

	version, err := dec.UnpackString()
	if err != nil {
		return "", err
	}
	abi.Version = version

	length, err := dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}
	for ; length > 0; length -= 1 {
		t := &ABIType{}
		t.NewTypeName, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		t.Type, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		abi.Types = append(abi.Types, *t)
	}

	length, err = dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}
	for ; length > 0; length -= 1 {
		s := &ABIStruct{}
		s.Name, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		s.Base, err = dec.UnpackString()
		if err != nil {
			return "", err
		}

		length2, err := dec.UnpackVarUint32()
		if err != nil {
			return "", err
		}
		for ; length2 > 0; length2 -= 1 {
			f := &ABIStructField{}
			f.Name, err = dec.UnpackString()
			if err != nil {
				return "", err
			}
			f.Type, err = dec.UnpackString()
			if err != nil {
				return "", err
			}
			s.Fields = append(s.Fields, *f)
		}
		abi.Structs = append(abi.Structs, *s)
	}

	length, err = dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}
	for ; length > 0; length -= 1 {
		a := &ABIAction{}
		name, err := dec.UnpackName()
		if err != nil {
			return "", err
		}
		a.Name = name.String()

		a.Type, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		a.RicardianContract, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		abi.Actions = append(abi.Actions, *a)
	}

	length, err = dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}
	for ; length > 0; length -= 1 {
		t := &ABITable{}
		name, err := dec.UnpackName()
		if err != nil {
			return "", err
		}
		t.Name = name.String()
		t.IndexType, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		length2, err := dec.UnpackVarUint32()
		if err != nil {
			return "", err
		}
		for ; length2 > 0; length2 -= 1 {
			key, err := dec.UnpackString()
			if err != nil {
				return "", err
			}
			t.KeyNames = append(t.KeyNames, key)
		}
		length2, err = dec.UnpackVarUint32()
		if err != nil {
			return "", err
		}
		for ; length2 > 0; length2 -= 1 {
			s, err := dec.UnpackString()
			if err != nil {
				return "", err
			}
			t.KeyTypes = append(t.KeyTypes, s)
		}
		t.Type, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		abi.Tables = append(abi.Tables, *t)
	}

	length, err = dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}
	for ; length > 0; length -= 1 {
		a := ClausePair{}
		a.Id, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		a.Body, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		abi.RicardianClauses = append(abi.RicardianClauses, a)
	}

	length, err = dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}
	for ; length > 0; length -= 1 {
		errCode, err := dec.UnpackUint64()
		if err != nil {
			return "", err
		}
		errMsg, err := dec.UnpackString()
		if err != nil {
			return "", err
		}
		abi.ErrorMessages = append(abi.ErrorMessages, ErrorMessage{
			ErrorCode: errCode,
			ErrorMsg:  errMsg,
		})
	}

	length, err = dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}
	for ; length > 0; length -= 1 {
		a := AbiExtension{}
		typ, err := dec.UnpackUint16()
		if err != nil {
			return "", err
		}
		a.Type = typ

		ext, err := dec.UnpackBytes()
		if err != nil {
			return "", err
		}
		a.Extension = ext
		abi.AbiExtensions = append(abi.AbiExtensions, a)
	}

	length, err = dec.UnpackVarUint32()
	if err != nil {
		return "", err
	}

	for ; length > 0; length -= 1 {
		v := VariantDef{}
		v.Name, err = dec.UnpackString()
		if err != nil {
			return "", err
		}

		length2, err := dec.UnpackVarUint32()
		if err != nil {
			return "", err
		}

		for ; length2 > 0; length2 -= 1 {
			tp, err := dec.UnpackString()
			if err != nil {
				return "", err
			}
			v.Types = append(v.Types, tp)
		}
		abi.Variants = append(abi.Variants, v)
	}

	ret, err := json.Marshal(abi)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

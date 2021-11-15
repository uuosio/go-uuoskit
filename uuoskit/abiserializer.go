package uuoskit

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/orderedmap"
	secp256k1 "github.com/uuosio/go-secp256k1"
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

type ABIType struct {
	NewTypeName string `json:"new_type_name"`
	Type        string `json:"type"`
}

type ClausePair struct {
	Id   string `json:"id"`
	Body string `json:"body"`
}

type ErrorMessage struct {
	ErrorCode uint64 `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

// struct variant_def {
// 	std::string              name{};
// 	std::vector<std::string> types{};
//  };

type VariantDef struct {
	Name  string   `json:"name"`
	Types []string `json:"types"`
}

//std::vector<std::pair<uint16_t, std::vector<char>>>;
type AbiExtension struct {
	Type      uint16 `json:"type_name"`
	Extension []byte `json:"extension"`
}

type ABI struct {
	Version          string         `json:"version"`
	Types            []ABIType      `json:"types"`
	Structs          []ABIStruct    `json:"structs"`
	Actions          []ABIAction    `json:"actions"`
	Tables           []ABITable     `json:"tables"`
	RicardianClauses []ClausePair   `json:"ricardian_clauses"`
	ErrorMessages    []ErrorMessage `json:"error_messages"`
	AbiExtensions    []AbiExtension `json:"abi_extensions"`
	// Variants         []VariantDef   `json:"variants"`
}

type ABISerializer struct {
	abiMap         map[string]*ABIStruct
	baseTypeMap    map[string]bool
	contractAbiMap map[string]*ABI
	enc            *Encoder
	dec            *Decoder
	contractName   string
}

var (
	gSerializer *ABISerializer = nil
)

func GetABISerializer() *ABISerializer {
	if gSerializer != nil {
		return gSerializer
	}

	gSerializer = &ABISerializer{}
	gSerializer.enc = NewEncoder(1024 * 1024)

	gSerializer.abiMap = make(map[string]*ABIStruct)
	gSerializer.baseTypeMap = make(map[string]bool)
	gSerializer.contractAbiMap = make(map[string]*ABI)

	gSerializer.SetContractABI("eosio.token", []byte(eosioTokenAbi))

	for _, typeName := range baseTypes {
		gSerializer.baseTypeMap[typeName] = true
	}

	abi := ABI{}
	err := json.Unmarshal([]byte(baseABI), &abi)
	if err != nil {
		panic(err)
	}

	for i := range abi.Structs {
		s := &abi.Structs[i]
		gSerializer.abiMap[s.Name] = s
	}

	return gSerializer
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

func (t *ABISerializer) PackActionArgs(contractName, actionName string, args []byte) ([]byte, error) {
	t.contractName = contractName

	t.enc.buf = t.enc.buf[:0]

	m := make(map[string]JsonValue)
	err := json.Unmarshal(args, &m)
	if err != nil {
		return nil, newError(err)
	}

	abiStruct := t.GetActionStruct(contractName, actionName)
	if abiStruct == nil {
		return nil, newErrorf("abi struct not found for %s::%s", contractName, actionName)
	}

	err = t.PackAbiStruct(contractName, abiStruct, m)
	if err != nil {
		return nil, newError(err)
	}

	bs := t.enc.buf
	t.enc.Reset()
	ret := make([]byte, len(bs))
	copy(ret, bs)
	return ret, nil
}

func (t *ABISerializer) UnpackActionArgs(contractName string, actionName string, packedValue []byte) ([]byte, error) {
	abiStruct := t.GetActionStruct(contractName, actionName)
	if abiStruct == nil {
		return nil, newErrorf("unknown action %s::%s", contractName, actionName)
	}

	result, err := t.UnpackAbiStruct(contractName, abiStruct, packedValue)
	if result == nil {
		return nil, newError(err)
	}
	bs, err := json.Marshal(result)
	if err != nil {
		return nil, newError(err)
	}
	return bs, nil
}

func (t *ABISerializer) PackAbiStructByName(contractName string, structName string, args string) ([]byte, error) {
	t.enc.Reset()
	t.contractName = contractName
	m := make(map[string]JsonValue)
	err := json.Unmarshal([]byte(args), &m)
	if err != nil {
		return nil, newError(err)
	}

	s := t.GetAbiStruct(contractName, structName)
	if s == nil {
		return nil, newErrorf("abi struct %s not found in %s", structName, contractName)
	}

	err = t.PackAbiStruct(contractName, s, m)
	if err != nil {
		return nil, newError(err)
	}
	return t.enc.Bytes(), nil
}

func (t *ABISerializer) PackAbiType(contractName, abiType string, args []byte) ([]byte, error) {
	t.contractName = contractName

	t.enc.Reset()

	m := make(map[string]JsonValue)
	err := json.Unmarshal(args, &m)
	if err != nil {
		return nil, newError(err)
	}

	abiStruct := t.GetAbiStruct(contractName, abiType)
	if abiStruct == nil {
		return nil, newErrorf("abi struct not found for %s::%s", contractName, abiType)
	}

	err = t.PackAbiStruct(contractName, abiStruct, m)
	if err != nil {
		return nil, newError(err)
	}

	bs := t.enc.Bytes()
	t.enc.Reset()
	ret := make([]byte, len(bs))
	copy(ret, bs)
	return ret, nil
}

func (t *ABISerializer) UnpackAbiType(contractName string, abiName string, packedValue []byte) ([]byte, error) {
	abiStruct := t.GetAbiStruct(contractName, abiName)
	if abiStruct == nil {
		return nil, newErrorf("unknown action %s::%s", contractName, abiName)
	}

	result, err := t.UnpackAbiStruct(contractName, abiStruct, packedValue)
	if result == nil {
		return nil, newError(err)
	}
	bs, err := json.Marshal(result)
	if err != nil {
		return nil, newError(err)
	}
	return bs, nil
}

func StringToInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, newError(err)
	}
	return i, nil
}

func StripString(v string) (string, bool) {
	v, err := strconv.Unquote(v)
	if err != nil {
		return "", false
	}
	return v, true
}

func IsSymbolValid(sym string) bool {
	if len(sym) > 7 || len(sym) <= 0 {
		return false
	}
	_vv := []byte(sym)
	i := 0
	for ; i < len(_vv); i++ {
		c := _vv[i]
		if c >= byte('A') && c <= byte('Z') {
		} else if c == 0 {
			for ; i < len(_vv); i++ {
				if _vv[i] != 0 {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

//{"quantity":"1.0000 EOS","contract":"eosio.token"}
type AbiExtendedAsset struct {
	Quantity string `json:"quantity"`
	Contract string `json:"contract"`
}

func (t *ABISerializer) ParseAbiStringValue(typ string, v string) error {
	switch typ {
	case "bool":
		if v == "true" || v == "1" {
			t.enc.PackBool(true)
		} else if v == "false" || v == "0" {
			t.enc.PackBool(false)
		} else {
			return newErrorf("invalid bool value: %s", v)
		}
		break
	case "int8":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt8 || n < math.MinInt8 {
			return newErrorf("int8 overflow: %d", n)
		}
		t.enc.PackInt8(int8(n))
		break
	case "uint8":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint8 || n < 0 {
			return newErrorf("uint8 overflow: %d", n)
		}
		t.enc.PackUint8(uint8(n))
		break
	case "int16":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt16 || n < math.MinInt16 {
			return newErrorf("int16 overflow: %d", n)
		}
		t.enc.PackInt16(int16(n))
		break
	case "uint16":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint16 || n < 0 {
			return newErrorf("uint16 overflow: %d", n)
		}
		t.enc.PackUint16(uint16(n))
		break
	case "int32":
		// log.Println("++++++++++int32:", v)
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt32 || n < math.MinInt32 {
			return newErrorf("int32 overflow: %d", n)
		}
		t.enc.PackInt32(int32(n))
	case "uint32":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint32 || n < 0 {
			return newErrorf("uint32 overflow: %d", n)
		}
		t.enc.PackUint32(uint32(n))
	case "int64":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt64 || n < math.MinInt64 {
			return newErrorf("int64 overflow: %d", n)
		}
		t.enc.PackInt64(int64(n))
		break
	case "uint64":
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint64 || n < 0 {
			return newErrorf("uint64 overflow: %d", n)
		}
		t.enc.PackUint64(uint64(n))
	case "int128", "uint128", "float128":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid %s, value: %s", typ, v)
		}

		if v[:2] != "0x" {
			return newErrorf("invalid %s, value: %s", typ, v)
		}

		v = v[2:]
		if len(v) != 32 {
			return newErrorf("invalid %s, %s, should be 0x followed by 32 hex character", typ, v)
		}

		bs, err := hex.DecodeString(v)
		if err != nil {
			return newError(err)
		}
		if len(bs) > 16 {
			return newErrorf("invalid int128 value: %s", v)
		}
		buf := make([]byte, 16)
		copy(buf[:], bs)
		t.enc.WriteBytes(buf)
	case "varint32":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt32 || n < math.MinInt32 {
			return newErrorf("varint32 overflow: %d", n)
		}
		t.enc.PackVarInt32(int32(n))
		break
	case "varuint32":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint32 || n < 0 {
			return newErrorf("varuint32 overflow: %d", n)
		}
		t.enc.PackVarUint32(uint32(n))
	case "float32":
		n, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return newError(err)
		}
		t.enc.PackFloat32(float32(n))
	case "float64":
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return newError(err)
		}
		t.enc.PackFloat64(n)
	case "time_point":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid time_point value: %s", v)
		}
		tt, err := time.Parse(time.RFC3339, string(v)+"Z")
		if err != nil {
			return newError(err)
		}
		n := tt.UnixNano()
		t.enc.PackInt64(n)
	case "time_point_sec", "block_timestamp_type":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid time_point_sec value: %s", v)
		}
		tt, err := time.Parse(time.RFC3339, string(v)+"Z")
		if err != nil {
			return newError(err)
		}
		n := tt.Unix()
		t.enc.PackUint32(uint32(n))
	case "name":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid name value: %s", v)
		}
		n := S2N(v)
		if N2S(n) != v {
			return newErrorf("invalid name value: %s", v)
		}
		t.enc.PackUint64(n)
	case "bytes":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid bytes value: %s", v)
		}
		bs, err := hex.DecodeString(v)
		if err != nil {
			return newError(err)
		}
		t.enc.PackBytes(bs)
	case "string":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid string value: %s", v)
		}
		t.enc.PackString(v)
	case "checksum160":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid checksum160 value: %s", v)
		}
		bs, err := hex.DecodeString(v)
		if err != nil {
			return newError(err)
		}
		if len(bs) > 20 {
			return newErrorf("invalid checksum160 value: %s", v)
		}
		buf := make([]byte, 20)
		copy(buf[:], bs)
		t.enc.WriteBytes(buf)
	case "checksum256":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid checksum256 value: %s", v)
		}
		bs, err := hex.DecodeString(v)
		if err != nil {
			return newError(err)
		}
		if len(bs) > 32 {
			return newErrorf("wrong checksum256 size: %s", v)
		}
		buf := make([]byte, 32)
		copy(buf[:], bs)
		t.enc.WriteBytes(buf)
	case "checksum512":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid checksum512 value: %s", v)
		}
		bs, err := hex.DecodeString(v)
		if err != nil {
			return newError(err)
		}
		if len(bs) > 64 {
			return newErrorf("invalid checksum512 value: %s", v)
		}
		buf := make([]byte, 64)
		copy(buf[:], bs)
		t.enc.WriteBytes(buf)
	case "public_key":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid public_key value: %s", v)
		}
		pub, err := secp256k1.NewPublicKeyFromBase58(v)
		if err != nil {
			return newError(err)
		}
		t.enc.WriteBytes([]byte{0})
		t.enc.WriteBytes(pub.Data[:])
		break
	case "signature":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid signature value: %s", v)
		}
		sig, err := secp256k1.NewSignatureFromBase58(v)
		if err != nil {
			return newError(err)
		}
		t.enc.WriteBytes([]byte{0})
		t.enc.WriteBytes(sig.Data[:])
	case "symbol":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid symbol value: %s", v)
		}
		vv := strings.Split(v, ",")
		if len(vv) != 2 {
			return newErrorf("invalid symbol value: %s", v)
		}
		n, err := strconv.ParseUint(vv[0], 10, 64)
		if err != nil {
			return newError(err)
		}
		if n > 16 {
			return newErrorf("invalid symbol value: %s", v)
		}
		if len(vv[1]) > 7 || len(vv[1]) <= 0 {
			return newErrorf("invalid symbol value: %s", v)
		}
		if !IsSymbolValid(vv[1]) {
			return newErrorf("invalid symbol value: %s", v)
		}

		_vv := []byte(vv[1])
		symbol_code := make([]byte, 8)
		symbol_code[0] = byte(n)
		copy(symbol_code[1:], _vv[:])
		t.enc.WriteBytes(symbol_code)
		break
	case "symbol_code":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid symbol_code value: %s", v)
		}
		if !IsSymbolValid(v) {
			return newErrorf("invalid symbol_code value: %s", v)
		}
		symbol_code := make([]byte, 8)
		copy(symbol_code[:], []byte(v))
		t.enc.WriteBytes(symbol_code)
	case "asset":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("StripString: Invalid asset value: %s", v)
		}
		r, ok := ParseAsset(v)
		if !ok {
			return newErrorf("ParseAsset: Invalid asset value: %s", v)
		}
		t.enc.WriteBytes(r)
	case "extended_asset":
		a := AbiExtendedAsset{}
		err := json.Unmarshal([]byte(v), &a)
		if err != nil {
			return newError(err)
		}
		r, ok := ParseAsset(a.Quantity)
		if !ok {
			return newErrorf("invalid asset value: %s", v)
		}
		t.enc.WriteBytes(r)

		n := S2N(a.Contract)
		if N2S(n) != v {
			return newErrorf("invalid name value: %s", v)
		}
		t.enc.PackUint64(n)
	default:
		return newErrorf("unsupported type: %s %T, %v\n", typ, v, v)
	}

	return nil
}

func (t *ABISerializer) unpackAbiStructField(typ string) (interface{}, error) {
	switch typ {
	case "bool":
		v, err := t.dec.UnpackBool()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int8":
		v, err := t.dec.UnpackInt8()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint8":
		v, err := t.dec.UnpackUint8()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int16":
		v, err := t.dec.UnpackInt16()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint16":
		v, err := t.dec.UnpackUint16()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int32":
		v, err := t.dec.UnpackInt32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint32":
		v, err := t.dec.UnpackUint32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int64":
		v, err := t.dec.UnpackInt64()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint64":
		v, err := t.dec.UnpackUint64()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int128", "uint128", "float128":
		buf := [16]byte{}
		err := t.dec.Read(buf[:])
		if err != nil {
			return nil, newError(err)
		}
		return "0x" + hex.EncodeToString(buf[:]), nil
	case "varint32":
		v, err := t.dec.UnpackVarInt32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "varuint32":
		v, err := t.dec.UnpackVarUint32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "float32":
		v, err := t.dec.UnpackFloat32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "float64":
		v, err := t.dec.UnpackFloat64()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "time_point":
		v, err := t.dec.ReadUint64()
		if err != nil {
			return nil, newError(err)
		}
		//convert seconds to iso8601
		return time.Unix(0, int64(v)).Format("2006-01-02T15:04:05"), nil
	case "time_point_sec", "block_timestamp_type":
		v, err := t.dec.ReadUint32()
		if err != nil {
			return nil, newError(err)
		}
		//convert seconds to iso8601
		return time.Unix(int64(v), 0).Format("2006-01-02T15:04:05"), nil
	case "name":
		v, err := t.dec.ReadUint64()
		if err != nil {
			return nil, newError(err)
		}
		return N2S(v), nil
	case "bytes":
		v, err := t.dec.UnpackBytes()
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "string":
		v, err := t.dec.UnpackString()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "checksum160":
		v := make([]byte, 20)
		err := t.dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "checksum256":
		v := make([]byte, 32)
		err := t.dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "checksum512":
		v := make([]byte, 64)
		err := t.dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "public_key":
		v := make([]byte, 34)
		err := t.dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		pub := secp256k1.PublicKey{}
		copy(pub.Data[:], v[1:])
		return pub.String(), nil
	case "signature":
		v := make([]byte, 66)
		err := t.dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		sig := secp256k1.Signature{}
		copy(sig.Data[:], v[1:])
		return sig.String(), nil
	case "symbol":
		buf := make([]byte, 8)
		err := t.dec.Read(buf)
		if err != nil {
			return nil, newError(err)
		}
		precision := int(buf[0])
		sym := string(buf[1:])
		sym = strings.TrimRight(sym, "\x00")
		return fmt.Sprintf("%s,%d", sym, precision), nil
	case "symbol_code":
		buf := make([]byte, 8)
		err := t.dec.Read(buf)
		if err != nil {
			return nil, newError(err)
		}
		sym := string(buf)
		sym = strings.TrimRight(sym, "\x00")
		return sym, nil
	case "asset":
		amount, err := t.dec.UnpackInt64()
		if err != nil {
			return nil, newError(err)
		}
		sym := make([]byte, 8)
		err = t.dec.Read(sym)
		if err != nil {
			return nil, newError(err)
		}
		precision := int64(sym[0])
		_sym := strings.TrimRight(string(sym[1:]), "\x00")
		format := fmt.Sprintf("%%d.%%0%dd %%s", precision)
		precision = int64(math.Pow10(int(precision)))
		return fmt.Sprintf(format, amount/precision, amount%precision, _sym), nil
	case "extended_asset":
		// {"quantity":"1.0000 EOS","contract":"eosio.token"}
		quantity, err := t.unpackAbiStructField("asset")
		if err != nil {
			return nil, newError(err)
		}
		contract, err := t.unpackAbiStructField("name")
		if err != nil {
			return nil, newError(err)
		}
		m := orderedmap.New()
		m.Set("quantity", quantity)
		m.Set("contract", contract)
		return m, nil
	default:
		return nil, newErrorf("unknown type %s", typ)
	}
}

func ParseAsset(v string) ([]byte, bool) {
	vv := strings.Split(v, " ")
	if len(vv) != 2 {
		return nil, false
	}
	amount := vv[0]
	symbol_code := vv[1]
	if !IsSymbolValid(symbol_code) {
		return nil, false
	}
	_, err := strconv.ParseFloat(amount, 64)

	if err != nil {
		return nil, false
	}
	_amount := strings.Split(amount, ".")
	if len(_amount) != 2 {
		return nil, false
	}

	// n, err := strconv.Atoi(_amount[1])
	// if err != nil {
	// 	return nil, false
	// }
	// log.Println("+++++ok here")
	// if n != 0 {
	// 	return nil, false
	// }
	precision := len(_amount[1])
	// __amount, err := strconv.Atoi(_amount[0])
	// if err != nil {
	// 	return nil, false
	// }
	// if __amount < 0 || __amount > math.MaxInt64 {
	// 	return nil, false
	// }

	amount = strings.Replace(amount, ".", "", 1)
	nAmount, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return nil, false
	}

	if nAmount < 0 || nAmount > math.MaxInt64 {
		return nil, false
	}

	enc := NewEncoder(8)
	enc.PackInt64(nAmount)

	enc.WriteBytes([]byte{byte(precision)})
	_symbol_code := make([]byte, 7)
	copy(_symbol_code[:], []byte(symbol_code)[:])
	enc.WriteBytes(_symbol_code)
	return enc.GetBytes(), true
}

func (t *ABISerializer) PackArrayAbiValue(typ string, value []JsonValue) error {
	t.enc.PackVarUint32(uint32(len(value)))
	for _, v := range value {
		_typ := strings.TrimSuffix(typ, "[]")
		err := t.ParseAbiValue(_typ, v)
		if err != nil {
			return newError(err)
		}
	}
	return nil
}

func (t *ABISerializer) GetBaseABIType(contractName string, typ string) (string, bool) {
	abi := t.contractAbiMap[contractName]
	if abi == nil {
		return "", false
	}
	for _, v := range abi.Types {
		if v.NewTypeName == typ {
			return v.Type, true
		}
	}
	return "", false
}

func (t *ABISerializer) PackAbiStruct(contractName string, abiStruct *ABIStruct, m map[string]JsonValue) error {
	for _, v := range abiStruct.Fields {
		typ := v.Type
		name := v.Name
		abiValue, ok := m[name]
		// log.Printf("++++++++PackAbiStruct: %s %s\n", name, typ)
		if !ok {
			//handle binary_extension
			if strings.HasSuffix(typ, "$") {
				continue
				//typ = strings.TrimSuffix(typ, "$")
			}
			return newErrorf("missing field %s", name)
		}

		if strings.HasSuffix(typ, "$") {
			typ = strings.TrimSuffix(typ, "$")
		} else if strings.HasSuffix(typ, "?") {
			typ = strings.TrimSuffix(typ, "?")
			if value, ok := abiValue.GetValue().(string); ok {
				if value == "null" {
					t.enc.PackBool(false)
					continue
				} else {
					t.enc.PackBool(true)
				}
			}
		}

		err := t.ParseAbiValue(typ, abiValue)
		if err == nil {
			continue
		}

		typ, ok = t.GetBaseABIType(contractName, typ)
		if !ok {
			return newError(err)
		}

		err = t.ParseAbiValue(typ, abiValue)
		if err != nil {
			return newError(err)
		}
	}
	return nil
}

func (t *ABISerializer) UnpackAbiStruct(contractName string, abiStruct *ABIStruct, packedValue []byte) (*orderedmap.OrderedMap, error) {
	t.dec = NewDecoder(packedValue)
	result := orderedmap.New()
	err := t.unpackAbiStruct(contractName, abiStruct, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *ABISerializer) unpackAbiStruct(contractName string, abiStruct *ABIStruct, result *orderedmap.OrderedMap) error {
	for _, v := range abiStruct.Fields {
		typ := v.Type
		name := v.Name

		//handle optional
		if strings.HasSuffix(typ, "?") {
			v, err := t.dec.UnpackBool()
			if err != nil {
				return newError(err)
			}
			if !v {
				result.Set(name, nil)
				continue
			}
			typ = strings.TrimRight(typ, "?")
		} else if strings.HasSuffix(typ, "$") { //handle binary_extension
			if t.dec.IsEnd() {
				return nil
			}
			typ = strings.TrimRight(typ, "$")
		}

		//try to find base type
		if baseName, ok := t.GetBaseName(contractName, typ); ok {
			typ = baseName
		}

		//try to unpack inner abi type
		if _, ok := t.baseTypeMap[typ]; ok {
			v, err := t.unpackAbiStructField(typ)
			if err == nil {
				result.Set(name, v)
				continue
			}
		}

		//try to unpack Abi struct
		subStruct := t.GetAbiStruct(contractName, typ)
		if subStruct != nil {
			subResult := orderedmap.New()
			result.Set(name, subResult)
			err := t.unpackAbiStruct(contractName, subStruct, subResult)
			if err != nil {
				return newError(err)
			}
			continue
		}

		//try to unpack array
		if !strings.HasSuffix(typ, "[]") {
			return newErrorf("unknown type %s", typ)
		}
		typ = strings.TrimSuffix(typ, "[]")
		arr := make([]interface{}, 0)
		count, err := t.dec.UnpackLength()
		if err != nil {
			return newError(err)
		}
		if _, ok := t.baseTypeMap[typ]; ok {
			for i := 0; i < count; i++ {
				v, err := t.unpackAbiStructField(typ)
				if err != nil {
					return newError(err)
				}
				arr = append(arr, v)
			}
			result.Set(name, arr)
			continue
		}

		subStruct = t.GetAbiStruct(contractName, typ)
		if subStruct == nil {
			return newErrorf("unknown type %s", typ)
		}
		for i := 0; i < count; i++ {
			subResult := orderedmap.New()
			err := t.unpackAbiStruct(contractName, subStruct, subResult)
			if err != nil {
				return newError(err)
			}
			arr = append(arr, subResult)
		}
		result.Set(name, arr)
	}
	return nil
}

func (t *ABISerializer) ParseAbiValue(typ string, abiValue JsonValue) error {
	switch v := abiValue.GetValue().(type) {
	case string:
		err := t.ParseAbiStringValue(typ, v)
		if err != nil {
			return newError(err)
		}
	case JsonValue:
		err := t.ParseAbiValue(typ, v)
		if err != nil {
			return newError(err)
		}
	case []JsonValue:
		err := t.PackArrayAbiValue(typ, v)
		if err != nil {
			return newError(err)
		}
	case map[string]JsonValue:
		s := t.GetAbiStruct(t.contractName, typ)
		if s == nil {
			return newErrorf("Unknown ABI type %s", typ)
		}
		err := t.PackAbiStruct(t.contractName, s, v)
		if err != nil {
			return newError(err)
		}
	default:
		return newErrorf("Unsupported type %[1]T: %[1]s", v)
	}
	return nil
}

func (t *ABISerializer) IsAbiCached(contractName string) bool {
	_, ok := t.contractAbiMap[contractName]
	return ok
}

func (t *ABISerializer) GetBaseName(contractName string, structName string) (string, bool) {
	abi := t.contractAbiMap[contractName]
	if abi == nil {
		return "", false
	}
	for j := range abi.Types {
		s := &abi.Types[j]
		if s.NewTypeName == structName {
			return s.Type, true
		}
	}
	return "", false
}

func (t *ABISerializer) GetAbiStruct(contractName string, structName string) *ABIStruct {
	abi := t.contractAbiMap[contractName]
	if abi == nil {
		return nil
	}
	for j := range abi.Structs {
		s := &abi.Structs[j]
		if s.Name == structName {
			return s
		}
	}
	s, ok := t.abiMap[structName]
	if !ok {
		return nil
	}
	return s
}

func (t *ABISerializer) GetActionStruct(contractName string, actionName string) *ABIStruct {
	abi, ok := t.contractAbiMap[contractName]
	if !ok {
		return nil
	}

	if abi.Actions == nil {
		return nil
	}

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

func (t *ABISerializer) GetActionStructName(contractName string, actionName string) string {
	abi := t.contractAbiMap[contractName]
	for i := range abi.Actions {
		action := &abi.Actions[i]
		if action.Name == actionName {
			return action.Type
		}
	}
	return ""
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
		a := &ErrorMessage{}
		a.ErrorCode, err = dec.UnpackUint64()
		if err != nil {
			return "", err
		}
		a.ErrorMsg, err = dec.UnpackString()
		if err != nil {
			return "", err
		}
		abi.ErrorMessages = append(abi.ErrorMessages, *a)
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

	ret, err := json.Marshal(abi)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

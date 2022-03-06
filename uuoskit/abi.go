package uuoskit

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
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

//{"quantity":"1.0000 EOS","contract":"eosio.token"}
type AbiExtendedAsset struct {
	Quantity string `json:"quantity"`
	Contract string `json:"contract"`
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
	Variants         []VariantDef   `json:"variants"`
}

func (t *ABI) PackAbiType(abiType string, args []byte) ([]byte, error) {
	enc := NewEncoder(1024)

	m := make(map[string]JsonValue)
	err := json.Unmarshal(args, &m)
	if err != nil {
		return nil, newError(err)
	}

	err = t.PackAbiStruct(enc, abiType, m)
	if err != nil {
		return nil, newError(err)
	}

	return enc.GetBytes(), nil
}

func (t *ABI) UnpackAbiType(abiName string, packedValue []byte) ([]byte, error) {
	dec := NewDecoder(packedValue)
	result := orderedmap.New()
	err := t.UnpackAbiStruct(dec, abiName, result)
	if err != nil {
		return nil, newError(err)
	}
	bs, err := json.Marshal(result)
	if err != nil {
		return nil, newError(err)
	}
	return bs, nil
}

func (t *ABI) PackAbiStructByName(structName string, args string) ([]byte, error) {
	enc := NewEncoder(1024)
	m := make(map[string]JsonValue)
	err := json.Unmarshal([]byte(args), &m)
	if err != nil {
		return nil, newError(err)
	}

	err = t.PackAbiStruct(enc, structName, m)
	if err != nil {
		return nil, newError(err)
	}
	return enc.Bytes(), nil
}

func (t *ABI) ParseAbiStringValue(enc *Encoder, typ string, v string) error {
	switch typ {
	case "bool":
		if v == "true" || v == "1" {
			enc.PackBool(true)
		} else if v == "false" || v == "0" {
			enc.PackBool(false)
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
		enc.PackInt8(int8(n))
		break
	case "uint8":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint8 || n < 0 {
			return newErrorf("uint8 overflow: %d", n)
		}
		enc.PackUint8(uint8(n))
		break
	case "int16":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt16 || n < math.MinInt16 {
			return newErrorf("int16 overflow: %d", n)
		}
		enc.PackInt16(int16(n))
		break
	case "uint16":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint16 || n < 0 {
			return newErrorf("uint16 overflow: %d", n)
		}
		enc.PackUint16(uint16(n))
		break
	case "int32":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt32 || n < math.MinInt32 {
			return newErrorf("int32 overflow: %d", n)
		}
		enc.PackInt32(int32(n))
	case "uint32":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint32 || n < 0 {
			return newErrorf("uint32 overflow: %d", n)
		}
		enc.PackUint32(uint32(n))
	case "int64":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt64 || n < math.MinInt64 {
			return newErrorf("int64 overflow: %d", n)
		}
		enc.PackInt64(int64(n))
		break
	case "uint64":
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint64 || n < 0 {
			return newErrorf("uint64 overflow: %d", n)
		}
		enc.PackUint64(uint64(n))
	case "int128", "uint128", "float128":
		if vv, ok := StripString(v); ok {
			if vv[:2] != "0x" {
				return newErrorf("invalid %s, value: %s", typ, vv)
			}

			vv = vv[2:]
			if typ == "float128" && len(vv) != 32 {
				return newErrorf("invalid %s, %s, should be 0x followed by 32 hex character", typ, vv)
			}

			bs, err := hex.DecodeString(vv)
			if err != nil {
				return newError(err)
			}
			if len(bs) > 16 {
				return newErrorf("invalid %s, value: %s", typ, v)
			}
			buf := make([]byte, 16)
			if typ != "float128" {
				reverseBytes(bs)
			}
			copy(buf[:], bs)
			enc.WriteBytes(buf)
		} else {
			if typ == "float128" {
				return newErrorf("invalid %s, value: %s", typ, v)
			}
			n := new(big.Int)
			n, ok := n.SetString(v, 10)
			if !ok {
				return newErrorf("invalid %s, value: %s", typ, v)
			}
			bs := n.Bytes()
			if len(bs) > 16 {
				return newErrorf("invalid %s, value: %s", typ, v)
			}
			if n.Sign() < 0 {
				if typ == "uint128" || typ == "float128" {
					return newErrorf("invalid %s, value: %s", typ, v)
				}
				m := new(big.Int)
				m.SetBytes([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				m.Sub(m, n.Abs(n))
				tmp := &big.Int{}
				tmp.SetUint64(1)
				m.Add(m, tmp)
				bs = m.Bytes()
			}
			buf := make([]byte, 16)
			reverseBytes(bs)
			copy(buf[:], bs)
			enc.WriteBytes(buf)
		}
	case "varint32":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxInt32 || n < math.MinInt32 {
			return newErrorf("varint32 overflow: %d", n)
		}
		enc.PackVarInt32(int32(n))
		break
	case "varuint32":
		n, err := StringToInt(v)
		if err != nil {
			return newError(err)
		}
		if n > math.MaxUint32 || n < 0 {
			return newErrorf("varuint32 overflow: %d", n)
		}
		enc.PackVarUint32(uint32(n))
	case "float32":
		n, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return newError(err)
		}
		enc.PackFloat32(float32(n))
	case "float64":
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return newError(err)
		}
		enc.PackFloat64(n)
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
		enc.PackInt64(n)
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
		enc.PackUint32(uint32(n))
	case "name":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid name value: %s", v)
		}
		n := S2N(v)
		if N2S(n) != v {
			return newErrorf("invalid name value: %s", v)
		}
		enc.PackUint64(n)
	case "bytes":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid bytes value: %s", v)
		}
		bs, err := hex.DecodeString(v)
		if err != nil {
			return newError(err)
		}
		enc.PackBytes(bs)
	case "string":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid string value: %s", v)
		}
		enc.PackString(v)
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
		enc.WriteBytes(buf)
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
		enc.WriteBytes(buf)
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
		enc.WriteBytes(buf)
	case "public_key":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("invalid public_key value: %s", v)
		}
		pub, err := secp256k1.NewPublicKeyFromBase58(v)
		if err != nil {
			return newError(err)
		}
		enc.WriteBytes([]byte{0})
		enc.WriteBytes(pub.Data[:])
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
		enc.WriteBytes([]byte{0})
		enc.WriteBytes(sig.Data[:])
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
		enc.WriteBytes(symbol_code)
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
		enc.WriteBytes(symbol_code)
	case "asset":
		v, ok := StripString(v)
		if !ok {
			return newErrorf("StripString: Invalid asset value: %s", v)
		}
		r, ok := ParseAsset(v)
		if !ok {
			return newErrorf("ParseAsset: Invalid asset value: %s", v)
		}
		enc.WriteBytes(r)
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
		enc.WriteBytes(r)

		n := S2N(a.Contract)
		if N2S(n) != v {
			return newErrorf("invalid name value: %s", v)
		}
		enc.PackUint64(n)
	default:
		return newErrorf("unsupported type: %s %T, %v\n", typ, v, v)
	}

	return nil
}

func (t *ABI) unpackAbiStructField(dec *Decoder, typ string) (interface{}, error) {
	switch typ {
	case "bool":
		v, err := dec.UnpackBool()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int8":
		v, err := dec.UnpackInt8()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint8":
		v, err := dec.UnpackUint8()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int16":
		v, err := dec.UnpackInt16()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint16":
		v, err := dec.UnpackUint16()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int32":
		v, err := dec.UnpackInt32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint32":
		v, err := dec.UnpackUint32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int64":
		v, err := dec.UnpackInt64()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "uint64":
		v, err := dec.UnpackUint64()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "int128", "uint128", "float128":
		buf := [16]byte{}
		err := dec.Read(buf[:])
		if err != nil {
			return nil, newError(err)
		}
		if typ != "float128" {
			reverseBytes(buf[:])
		}
		return "0x" + hex.EncodeToString(buf[:]), nil
	case "varint32":
		v, err := dec.UnpackVarInt32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "varuint32":
		v, err := dec.UnpackVarUint32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "float32":
		v, err := dec.UnpackFloat32()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "float64":
		v, err := dec.UnpackFloat64()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "time_point":
		v, err := dec.ReadUint64()
		if err != nil {
			return nil, newError(err)
		}
		//convert seconds to iso8601
		return time.Unix(0, int64(v)).Format("2006-01-02T15:04:05"), nil
	case "time_point_sec", "block_timestamp_type":
		v, err := dec.ReadUint32()
		if err != nil {
			return nil, newError(err)
		}
		//convert seconds to iso8601
		return time.Unix(int64(v), 0).Format("2006-01-02T15:04:05"), nil
	case "name":
		v, err := dec.ReadUint64()
		if err != nil {
			return nil, newError(err)
		}
		return N2S(v), nil
	case "bytes":
		v, err := dec.UnpackBytes()
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "string":
		v, err := dec.UnpackString()
		if err != nil {
			return nil, newError(err)
		}
		return v, nil
	case "checksum160":
		v := make([]byte, 20)
		err := dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "checksum256":
		v := make([]byte, 32)
		err := dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "checksum512":
		v := make([]byte, 64)
		err := dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		return hex.EncodeToString(v), nil
	case "public_key":
		v := make([]byte, 34)
		err := dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		pub := secp256k1.PublicKey{}
		copy(pub.Data[:], v[1:])
		return pub.StringEOS(), nil
	case "signature":
		v := make([]byte, 66)
		err := dec.Read(v)
		if err != nil {
			return nil, newError(err)
		}
		sig := secp256k1.Signature{}
		copy(sig.Data[:], v[1:])
		return sig.String(), nil
	case "symbol":
		buf := make([]byte, 8)
		err := dec.Read(buf)
		if err != nil {
			return nil, newError(err)
		}
		precision := int(buf[0])
		sym := string(buf[1:])
		sym = strings.TrimRight(sym, "\x00")
		return fmt.Sprintf("%s,%d", sym, precision), nil
	case "symbol_code":
		buf := make([]byte, 8)
		err := dec.Read(buf)
		if err != nil {
			return nil, newError(err)
		}
		sym := string(buf)
		sym = strings.TrimRight(sym, "\x00")
		return sym, nil
	case "asset":
		amount, err := dec.UnpackInt64()
		if err != nil {
			return nil, newError(err)
		}
		sym := make([]byte, 8)
		err = dec.Read(sym)
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
		quantity, err := t.unpackAbiStructField(dec, "asset")
		if err != nil {
			return nil, newError(err)
		}
		contract, err := t.unpackAbiStructField(dec, "name")
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

func (t *ABI) PackArrayAbiValue(enc *Encoder, typ string, value []JsonValue) error {
	enc.PackVarUint32(uint32(len(value)))
	for _, v := range value {
		_typ := strings.TrimSuffix(typ, "[]")
		err := t.PackAbiValue(enc, _typ, v)
		if err != nil {
			return newError(err)
		}
	}
	return nil
}

func (t *ABI) GetBaseABIType(typ string) (string, bool) {
	for _, v := range t.Types {
		if v.NewTypeName == typ {
			return v.Type, true
		}
	}
	return "", false
}

func (t *ABI) GetVariantType(typ string) (*VariantDef, bool) {
	for i := range t.Variants {
		v := &t.Variants[i]
		if v.Name == typ {
			return v, true
		}
	}
	return nil, false
}

func (t *ABI) PackAbiStruct(enc *Encoder, structName string, m map[string]JsonValue) error {
	abiStruct := t.GetAbiStruct(structName)
	if abiStruct == nil {
		return newErrorf("abi struct %s not found", structName)
	}

	if abiStruct.Base != "" {
		err := t.PackAbiStruct(enc, abiStruct.Base, m)
		if err != nil {
			return newError(err)
		}
	}

	for _, v := range abiStruct.Fields {
		typ := v.Type
		name := v.Name
		abiValue, ok := m[name]
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
					enc.PackBool(false)
					continue
				} else {
					enc.PackBool(true)
				}
			}
		}

		err := t.PackAbiValue(enc, typ, abiValue)
		if err == nil {
			continue
		}

		typ, ok = t.GetBaseABIType(typ)
		if !ok {
			return newError(err)
		}

		err = t.PackAbiValue(enc, typ, abiValue)
		if err != nil {
			return newError(err)
		}
	}
	return nil
}

func (t *ABI) UnpackAbiStruct(dec *Decoder, actionType string, result *orderedmap.OrderedMap) error {
	abiStruct := t.GetAbiStruct(actionType)
	if abiStruct == nil {
		return newErrorf("abi struct %s not found", actionType)
	}

	if abiStruct.Base != "" {
		err := t.UnpackAbiStruct(dec, abiStruct.Base, result)
		if err != nil {
			return newError(err)
		}
	}

	err := t.unpackAbiStructFields(dec, abiStruct.Fields, result)
	if err != nil {
		return err
	}
	return nil
}

func (t *ABI) unpackAbiStructFields(dec *Decoder, fields []ABIStructField, result *orderedmap.OrderedMap) error {
	for _, v := range fields {
		typ := v.Type
		name := v.Name

		//handle optional
		if strings.HasSuffix(typ, "?") {
			v, err := dec.UnpackBool()
			if err != nil {
				return newError(err)
			}
			if !v {
				result.Set(name, nil)
				continue
			}
			typ = strings.TrimRight(typ, "?")
		} else if strings.HasSuffix(typ, "$") { //handle binary_extension
			if dec.IsEnd() {
				return nil
			}
			typ = strings.TrimRight(typ, "$")
		}

		//try to find base type
		if baseName, ok := t.GetBaseName(typ); ok {
			typ = baseName
		}
		//try to unpack inner abi type
		if _, ok := gBaseTypes[typ]; ok {
			v, err := t.unpackAbiStructField(dec, typ)
			if err != nil {
				return err
			}
			result.Set(name, v)
			continue
		}

		//try to unpack Abi struct
		subStruct := t.GetAbiStruct(typ)
		if subStruct != nil {
			subResult := orderedmap.New()
			result.Set(name, subResult)
			err := t.UnpackAbiStruct(dec, typ, subResult)
			if err != nil {
				return newError(err)
			}
			continue
		}
		//try to unpack variant type
		if v, ok := t.GetVariantType(typ); ok {
			index, err := dec.UnpackUint8()
			if err != nil {
				return err
			}

			if int(index) >= len(v.Types) {
				return newErrorf("invalid variant index %d", index)
			}
			tp := v.Types[int(index)]
			arr := make([]interface{}, 0)
			arr = append(arr, tp)
			value, err := t.unpackAbiStructField(dec, tp)
			if err != nil {
				return err
			}
			arr = append(arr, value)
			result.Set(name, arr)
			continue
		}
		//try to unpack array
		if !strings.HasSuffix(typ, "[]") {
			return newErrorf("unknown type %s", typ)
		}
		typ = strings.TrimSuffix(typ, "[]")
		arr := make([]interface{}, 0)
		count, err := dec.UnpackLength()
		if err != nil {
			return newError(err)
		}

		//try to unpack inner abi type
		if _, ok := gBaseTypes[typ]; ok {
			for i := 0; i < count; i++ {
				v, err := t.unpackAbiStructField(dec, typ)
				if err != nil {
					return newError(err)
				}
				arr = append(arr, v)
			}
			result.Set(name, arr)
			continue
		}

		subStruct = t.GetAbiStruct(typ)
		if subStruct == nil {
			return newErrorf("unknown type %s", typ)
		}
		for i := 0; i < count; i++ {
			subResult := orderedmap.New()
			err := t.UnpackAbiStruct(dec, typ, subResult)
			if err != nil {
				return newError(err)
			}
			arr = append(arr, subResult)
		}
		result.Set(name, arr)
	}
	return nil
}

func (t *ABI) PackAbiValue(enc *Encoder, typ string, abiValue JsonValue) error {
	switch v := abiValue.GetValue().(type) {
	case string:
		err := t.ParseAbiStringValue(enc, typ, v)
		if err != nil {
			return newError(err)
		}
	case JsonValue:
		err := t.PackAbiValue(enc, typ, v)
		if err != nil {
			return newError(err)
		}
	case []JsonValue:
		if varType, ok := t.GetVariantType(typ); ok {
			if len(v) != 2 {
				return newErrorf("Invalid variant value %v", v)
			}
			innerType, ok := v[0].GetStringValue()
			if !ok {
				return newErrorf("+++++Invalid variant value %v", v)
			}
			found := false
			for i, variantType := range varType.Types {
				if variantType == innerType {
					enc.PackUint8(uint8(i))
					t.PackAbiValue(enc, variantType, v[1])
					found = true
					break
				}
			}
			if !found {
				return newErrorf("type %s not found in variant %v", innerType, typ)
			}
		} else {
			err := t.PackArrayAbiValue(enc, typ, v)
			if err != nil {
				return newError(err)
			}
		}
	case map[string]JsonValue:
		err := t.PackAbiStruct(enc, typ, v)
		if err != nil {
			return newError(err)
		}
	default:
		return newErrorf("Unsupported type %[1]T: %[1]s", v)
	}
	return nil
}

func (t *ABI) GetBaseName(structName string) (string, bool) {
	for j := range t.Types {
		s := &t.Types[j]
		if s.NewTypeName == structName {
			return s.Type, true
		}
	}
	return "", false
}

func (t *ABI) GetAbiStruct(structName string) *ABIStruct {
	for j := range t.Structs {
		s := &t.Structs[j]
		if s.Name == structName {
			return s
		}
	}
	return nil
}

func (t *ABI) GetActionStruct(actionName string) *ABIStruct {
	if t.Actions == nil {
		return nil
	}

	for i := range t.Actions {
		action := &t.Actions[i]
		if action.Name == actionName {
			for j := range t.Structs {
				s := &t.Structs[j]
				if s.Name == action.Type {
					return s
				}
			}
		}
	}
	return nil
}

func (t *ABI) GetActionStructType(actionName string) string {
	for i := range t.Actions {
		action := &t.Actions[i]
		if action.Name == actionName {
			return action.Type
		}
	}
	return ""
}

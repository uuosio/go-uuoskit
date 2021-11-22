package uuoskit

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type Bytes []byte

func (t Bytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString([]byte(t)))
}

func (a *Bytes) UnmarshalJSON(b []byte) error {
	bs, err := hex.DecodeString(strings.Trim(string(b), `"`))
	if err != nil {
		return newError(err)
	}
	*a = append((*a)[:0], bs...)
	return nil
}

type VarInt32 int32

func (t *VarInt32) Pack() []byte {
	return PackVarInt32(int32(*t))
}

func (t *VarInt32) Unpack(data []byte) (int, error) {
	v, n := UnpackVarInt32(data)
	*t = VarInt32(v)
	return n, nil
}

func (t *VarInt32) Size() int {
	return PackedVarInt32Length(int32(*t))
}

type VarUint32 uint32

func (t *VarUint32) Pack() []byte {
	return PackVarUint32(uint32(*t))
}

func (t *VarUint32) Unpack(data []byte) (int, error) {
	v, n := UnpackVarUint32(data)
	*t = VarUint32(v)
	return n, nil
}

func (t *VarUint32) Size() int {
	return PackedVarUint32Length(uint32(*t))
}

func (t *VarUint32) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(*t))
}

type Int128 [16]byte

func (n *Int128) Pack() []byte {
	return n[:]
}

func (n *Int128) Unpack(data []byte) (int, error) {
	dec := NewDecoder(data)
	if err := dec.Read(n[:]); err != nil {
		return 0, err
	}
	return 16, nil
}

func (t *Int128) Size() int {
	return 16
}

type Uint128 [16]byte

func (n *Uint128) Pack() []byte {
	return n[:]
}

func (n *Uint128) Unpack(data []byte) (int, error) {
	dec := NewDecoder(data)
	if err := dec.Read(n[:]); err != nil {
		return 0, err
	}
	return 16, nil
}

func (t *Uint128) Size() int {
	return 16
}

func (n *Uint128) SetUint64(v uint64) {
	tmp := Uint128{}
	copy(n[:], tmp[:]) //memset
	binary.LittleEndian.PutUint64(n[:], v)
}

func (n *Uint128) Uint64() uint64 {
	return binary.LittleEndian.Uint64(n[:])
}

type Uint256 [32]uint8

func (n *Uint256) Pack() []byte {
	return n[:]
}

func (n *Uint256) Unpack(data []byte) (int, error) {
	dec := NewDecoder(data)
	if err := dec.Read(n[:]); err != nil {
		return 0, err
	}
	return 32, nil
}

func (t *Uint256) Size() int {
	return 32
}

func (n *Uint256) SetUint64(v uint64) {
	tmp := Uint256{}
	copy(n[:], tmp[:]) //memset
	binary.LittleEndian.PutUint64(n[:], v)
}

func (n *Uint256) Uint64() uint64 {
	return binary.LittleEndian.Uint64(n[:])
}

type Float128 [16]byte

func (n *Float128) Pack() []byte {
	return n[:]
}

func (n *Float128) Unpack(data []byte) (int, error) {
	dec := NewDecoder(data)
	if err := dec.Read(n[:]); err != nil {
		return 0, err
	}
	return 16, nil
}

func (t *Float128) Size() int {
	return 16
}

type TimePoint struct {
	Elapsed uint64
}

func (t *TimePoint) Pack() []byte {
	enc := NewEncoder(t.Size())
	enc.PackUint64(t.Elapsed)
	return enc.GetBytes()
}

func (t *TimePoint) Unpack(data []byte) (int, error) {
	dec := NewDecoder(data)
	dec.Unpack(&t.Elapsed)
	return 8, nil
}

func (t *TimePoint) Size() int {
	return 8
}

type TimePointSec struct {
	UTCSeconds uint32
}

func (t *TimePointSec) Pack() []byte {
	enc := NewEncoder(t.Size())
	enc.PackUint32(t.UTCSeconds)
	return enc.GetBytes()
}

func (t *TimePointSec) Unpack(data []byte) (int, error) {
	dec := NewDecoder(data)
	dec.Unpack(&t.UTCSeconds)
	return 4, nil
}

func (t *TimePointSec) Size() int {
	return 4
}

func (t TimePointSec) MarshalJSON() ([]byte, error) {
	s := time.Unix(int64(t.UTCSeconds), 0).UTC().Format("2006-01-02T15:04:05")
	return json.Marshal(s)
}

func (a *TimePointSec) UnmarshalJSON(b []byte) error {
	t, err := time.Parse("2006-01-02T15:04:05", strings.Trim(string(b), "\""))
	if err != nil {
		return newError(err)
	}
	a.UTCSeconds = uint32(t.Unix())
	return nil
}

type BlockTimestampType struct {
	Slot uint32
}

func (t *BlockTimestampType) Pack() []byte {
	enc := NewEncoder(t.Size())
	enc.PackUint32(t.Slot)
	return enc.GetBytes()
}

func (t *BlockTimestampType) Unpack(data []byte) (int, error) {
	dec := NewDecoder(data)
	dec.Unpack(&t.Slot)
	return 4, nil
}

func (t *BlockTimestampType) Size() int {
	return 4
}

type JsonValue struct {
	value interface{}
}

func NewJsonValue(value interface{}) JsonValue {
	retValue := JsonValue{}
	retValue.SetValue(value)
	return retValue
}

func (b *JsonValue) GetValue() interface{} {
	return b.value
}

func (b *JsonValue) SetValue(value interface{}) error {
	switch value.(type) {
	case string, []JsonValue, map[string]JsonValue:
		b.value = value
		return nil
	default:
		panic("value must be a string, slice, or map")
	}
	return nil
}

func parseSubValue(subValue JsonValue) interface{} {
	switch v := subValue.value.(type) {
	case string:
		return strings.Trim(v, "\"")
	case []JsonValue, map[string]JsonValue:
		return subValue.value
	}
	return subValue
}

//return string, []JsonValue, or map[string]JsonValue
func (b *JsonValue) Get(keys ...interface{}) (interface{}, error) {
	if len(keys) == 0 {
		return JsonValue{}, newErrorf("no key specified")
	}

	value := b.value
	for _, key := range keys {
		switch v := key.(type) {
		case string:
			switch v2 := value.(type) {
			case map[string]JsonValue:
				subValue, ok := v2[v]
				if !ok {
					return nil, newErrorf("key not found")
				}
				value = parseSubValue(subValue)
			case JsonValue:
				v3, ok := v2.value.(map[string]JsonValue)
				if !ok {
					return JsonValue{}, newErrorf("1:JsonValue is not a map")
				}
				subValue, ok := v3[v]
				if !ok {
					return nil, newErrorf("key not found")
				}
				value = parseSubValue(subValue)
			default:
				return JsonValue{}, newErrorf("2:JsonValue is not a map")
			}
		case int:
			var arr []JsonValue
			var ok bool
			switch v := value.(type) {
			case []JsonValue:
				arr = v
			case JsonValue:
				arr, ok = v.value.([]JsonValue)
				if !ok {
					return JsonValue{}, newErrorf("JsonValue is not an array")
				}
			default:
				return JsonValue{}, newErrorf("JsonValue is not an array")
			}
			if v < 0 || v >= len(arr) {
				return JsonValue{}, newErrorf("index out of range")
			}
			value = arr[v]
		default:
			return JsonValue{}, newErrorf("invalid key type")
		}
	}
	return value, nil
}

func (b *JsonValue) GetString(keys ...interface{}) (string, error) {
	v, err := b.Get(keys...)
	if err != nil {
		return "", err
	}

	_v, ok := v.(string)
	if !ok {
		return "", newErrorf("value is not a string")
	}
	return strings.Trim(_v, "\""), nil
}

func (b *JsonValue) GetTime(keys ...interface{}) (*time.Time, error) {
	v, err := b.GetString(keys...)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-01-02T15:04:05", v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (b JsonValue) MarshalJSON() ([]byte, error) {
	switch v := b.value.(type) {
	case string:
		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			return []byte(v), nil
		} else {
			return []byte(strconv.Quote(v)), nil
		}
	case JsonValue:
		if s, ok := v.value.(string); ok {
			return []byte(s), nil
		}
		return json.Marshal(v.value)
	case map[string]JsonValue:
		return json.Marshal(v)
	case []JsonValue:
		return json.Marshal(v)
	}
	return nil, newErrorf("bad JsonValue")
}

func (b *JsonValue) UnmarshalJSON(data []byte) error {
	// fmt.Println("+++++:UnmarshaJSON", string(data))
	if data[0] == '{' {
		m := make(map[string]JsonValue)
		err := json.Unmarshal(data, &m)
		if err != nil {
			return newError(err)
		}
		b.value = m
	} else if data[0] == '[' {
		m := make([]JsonValue, 0, 1)
		err := json.Unmarshal(data, &m)
		if err != nil {
			return newError(err)
		}
		b.value = m
	} else {
		b.value = string(data)
	}
	return nil
}

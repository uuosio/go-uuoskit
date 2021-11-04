package uuoskit

import (
	"encoding/json"
	"testing"
)

func TestJsonValue(t *testing.T) {
	var rootJsonValue JsonValue
	jsonValue := NewJsonValue("hello,world")
	jsonValue2 := NewJsonValue("goodbye,world")
	jsonValue3 := NewJsonValue("123")

	jsonValueArray := []JsonValue{}
	jsonValueArray = append(jsonValueArray, jsonValue2)
	jsonValueArray = append(jsonValueArray, jsonValue3)

	jsonValueMap := map[string]JsonValue{}
	jsonValueMap["foo"] = NewJsonValue(jsonValueArray)
	jsonValueMap["bar"] = jsonValue

	rootJsonValue.SetValue(jsonValueMap)
	r, err := rootJsonValue.Get("foo", 0)
	if err != nil {
		t.Error(err)
	}
	{
		r2 := r.(JsonValue)
		t.Logf("+++%T  %v", r2, r2.GetValue().(string))
	}

	{
		r, err := json.Marshal(jsonValueMap)
		if err != nil {
			t.Error(err)
		}
		t.Logf(string(r))
	}
	{
		r, err := json.Marshal(rootJsonValue)
		if err != nil {
			t.Error(err)
		}
		t.Logf(string(r))
	}
}

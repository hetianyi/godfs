package json

import "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary


func Marshal(obj *interface{}) ([]byte, error) {
	return json.Marshal(&obj)
}

func Unmarshal(input []byte, obj *interface{}) error {
	return json.Unmarshal(input, obj)
}

func UnmarshalFromString(input string, obj *interface{}) error {
	return json.UnmarshalFromString(input, obj)
}
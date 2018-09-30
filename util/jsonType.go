package util

import (
	"encoding/json"
	"github.com/dailing/levlog"
	"reflect"
	"strconv"
)

type JsonType map[string]interface{}

func NewJson() *JsonType {
	retval := make(JsonType)
	return &retval
}
func NewJsonFromString(s string) *JsonType {
	retval := make(JsonType)
	err := json.Unmarshal([]byte(s), &retval)
	levlog.E(err)
	return &retval
}

func (j *JsonType) GetString(key string) string {
	if f, ok := (*j)[key].(string); ok {
		return f
	}
	if f, ok := (*j)[key].([]byte); ok {
		return string(f)
	}
	levlog.Error("Error convert ", key)
	return ""
}

func (j *JsonType) GetBytes(key string) []byte {
	if f, ok := (*j)[key].([]byte); ok {
		return f
	}
	if f, ok := (*j)[key].(string); ok {
		return []byte(f)
	}
	levlog.Error("Error convert :", key)
	return nil
}

func (j *JsonType) GetJson(key string) *JsonType {
	if f, ok := (*j)[key].(*JsonType); ok {
		return f
	}
	levlog.Warning("Not JsonType,", (*j)[key])
	payload, err := json.Marshal((*j)[key])
	levlog.E(err)
	return NewJsonFromString(string(payload))
}

func (j *JsonType) GetObj(key string) interface{} {
	return (*j)[key]
}

func (j *JsonType) GetInt(key string) int {
	val, ok := (*j)[key]
	if !ok {
		return 0
	}
	if s, ok := val.(string); ok {
		i, err := strconv.Atoi(s)
		levlog.E(err)
		return i
	}
	if i, ok := val.(int); ok {
		return i
	}
	if i, ok := val.(int64); ok {
		return int(i)
	}
	if i, ok := val.(float64); ok {
		return int(i)
	}
	if i, ok := val.(float32); ok {
		return int(i)
	}
	levlog.Error("Get Key Error ", reflect.TypeOf(val))
	return 0
}

func (j *JsonType) Set(key string, val interface{}) {
	(*j)[key] = val
}

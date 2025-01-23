package vigilant

import (
	"fmt"

	"go.opentelemetry.io/otel/log"
)

// Attribute is a map of metadata to be sent with the error
type Attribute struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// NewAttribute creates a new Attribute
func NewAttribute(key string, value interface{}) Attribute {
	return Attribute{
		Key:   key,
		Value: value,
	}
}

// toLogKV converts the attribute to a log.KeyValue
func (a Attribute) toLogKV() log.KeyValue {
	switch v := a.Value.(type) {
	case int:
		return log.Int(a.Key, v)
	case int64:
		return log.Int64(a.Key, v)
	case float64:
		return log.Float64(a.Key, v)
	case bool:
		return log.Bool(a.Key, v)
	case string:
		return log.String(a.Key, v)
	case []byte:
		return log.Bytes(a.Key, v)
	case float32:
		return log.Float64(a.Key, float64(v))
	case uint:
		return log.Int64(a.Key, int64(v))
	case uint64:
		return log.Int64(a.Key, int64(v))
	case uint32:
		return log.Int64(a.Key, int64(v))
	case uint16:
		return log.Int64(a.Key, int64(v))
	case uint8:
		return log.Int64(a.Key, int64(v))
	case uintptr:
		return log.Int64(a.Key, int64(v))
	case int8:
		return log.Int64(a.Key, int64(v))
	case int16:
		return log.Int64(a.Key, int64(v))
	case int32:
		return log.Int64(a.Key, int64(v))
	}
	return log.String(a.Key, fmt.Sprintf("%v", a.Value))
}

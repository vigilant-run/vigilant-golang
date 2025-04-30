package vigilant

// Tag represents a tag in an observability event.
type MetricTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// String returns the string representation of an attribute.
func Tag(key string, val string) MetricTag {
	return MetricTag{
		Key:   key,
		Value: val,
	}
}

package vigilant

// Attribute is a map of metadata to be sent with the error
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NewAttribute creates a new Attribute
func NewAttribute(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

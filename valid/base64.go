package valid

import "regexp"

var base64Pattern = regexp.MustCompile(`^(?:[A-Za-z0-99+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)

// Base64 check struct
type Base64 struct {
	Match
	Key string
}

// DefaultMessage return the default Base64 error message
func (b Base64) DefaultMessage() string {
	return MessageTmpfs["Base64"]
}

// GetKey return the b.Key
func (b Base64) GetKey() string {
	return b.Key
}

// GetLimitValue return the limit value
func (b Base64) GetLimitValue() interface{} {
	return nil
}

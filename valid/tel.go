package valid

import "regexp"

// just for chinese telephone number
var telPattern = regexp.MustCompile(`^(0\d{2,3}(\-)?)?\d{7,8}$`)

// Tel check telephone struct
type Tel struct {
	Match
	Key string
}

// DefaultMessage return the default Tel error message
func (t Tel) DefaultMessage() string {
	return MessageTmpfs["Tel"]
}

// GetKey return the t.Key
func (t Tel) GetKey() string {
	return t.Key
}

// GetLimitValue return the limit value
func (t Tel) GetLimitValue() interface{} {
	return nil
}

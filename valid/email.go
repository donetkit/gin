package valid

import "regexp"

var emailPattern = regexp.MustCompile(`^[\w!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[\w!#$%&'*+/=?^_` + "`" + `{|}~-]+)*@(?:[\w](?:[\w-]*[\w])?\.)+[a-zA-Z0-9](?:[\w-]*[\w])?$`)

// Email check struct
type Email struct {
	Match
	Key string
}

// DefaultMessage return the default Email error message
func (e Email) DefaultMessage() string {
	return MessageTmpfs["Email"]
}

// GetKey return the n.Key
func (e Email) GetKey() string {
	return e.Key
}

// GetLimitValue return the limit value
func (e Email) GetLimitValue() interface{} {
	return nil
}

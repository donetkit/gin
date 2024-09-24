package valid

import "regexp"

var ipPattern = regexp.MustCompile(`^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`)

// IP check struct
type IP struct {
	Match
	Key string
}

// DefaultMessage return the default IP error message
func (i IP) DefaultMessage() string {
	return MessageTmpfs["IP"]
}

// GetKey return the i.Key
func (i IP) GetKey() string {
	return i.Key
}

// GetLimitValue return the limit value
func (i IP) GetLimitValue() interface{} {
	return nil
}

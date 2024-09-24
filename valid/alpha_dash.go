package valid

import "regexp"

var alphaDashPattern = regexp.MustCompile(`[^\d\w-_]`)

// AlphaDash check not Alpha
type AlphaDash struct {
	NoMatch
	Key string
}

// DefaultMessage return the default AlphaDash error message
func (a AlphaDash) DefaultMessage() string {
	return MessageTmpfs["AlphaDash"]
}

// GetKey return the n.Key
func (a AlphaDash) GetKey() string {
	return a.Key
}

// GetLimitValue return the limit value
func (a AlphaDash) GetLimitValue() interface{} {
	return nil
}

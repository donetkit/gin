package valid

import "fmt"

// NoMatch Requires a string to not match a given regex.
type NoMatch struct {
	Match
	Key string
}

// IsSatisfied judge whether obj is valid
func (n NoMatch) IsSatisfied(obj interface{}) bool {
	return !n.Match.IsSatisfied(obj)
}

// DefaultMessage return the default NoMatch error message
func (n NoMatch) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["NoMatch"], n.Regexp.String())
}

// GetKey return the n.Key
func (n NoMatch) GetKey() string {
	return n.Key
}

// GetLimitValue return the limit value
func (n NoMatch) GetLimitValue() interface{} {
	return n.Regexp.String()
}

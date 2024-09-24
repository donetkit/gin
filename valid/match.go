package valid

import (
	"fmt"
	"regexp"
)

// Match Requires a string to match a given regex.
type Match struct {
	Regexp *regexp.Regexp
	Key    string
}

// IsSatisfied judge whether obj is valid
func (m Match) IsSatisfied(obj interface{}) bool {
	return m.Regexp.MatchString(fmt.Sprintf("%v", obj))
}

// DefaultMessage return the default Match error message
func (m Match) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Match"], m.Regexp.String())
}

// GetKey return the m.Key
func (m Match) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value
func (m Match) GetLimitValue() interface{} {
	return m.Regexp.String()
}

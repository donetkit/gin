package valid

import "fmt"

// Range Requires an integer to be within Min, Max inclusive.
type Range struct {
	Min
	Max
	Key string
}

// IsSatisfied judge whether obj is valid
// not support int64 on 32-bit platform
func (r Range) IsSatisfied(obj interface{}) bool {
	return r.Min.IsSatisfied(obj) && r.Max.IsSatisfied(obj)
}

// DefaultMessage return the default Range error message
func (r Range) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Range"], r.Min.Min, r.Max.Max)
}

// GetKey return the m.Key
func (r Range) GetKey() string {
	return r.Key
}

// GetLimitValue return the limit value, Max
func (r Range) GetLimitValue() interface{} {
	return []int{r.Min.Min, r.Max.Max}
}

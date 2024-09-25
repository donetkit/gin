package valid

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	// ValidTag struct tag
	ValidTag = "valid"

	LabelTag = "label"

	wordSize = 32 << (^uint(0) >> 32 & 1)
)

var (
	// key: function name
	// value: the number of parameters
	reflectFunc = make(ReflectFunc)

	// doesn't belong to validation functions
	unReflectFunc = map[string]bool{
		"Clear":     true,
		"HasErrors": true,
		"ErrorMap":  true,
		"Error":     true,
		"apply":     true,
		"Check":     true,
		"Valid":     true,
		"NoMatch":   true,
	}
	// ErrInt64On32 show 32-bit platform not support int64
	ErrInt64On32 = fmt.Errorf("not support int64 on 32-bit platform")
)

// ValidFormer valid interface
type ValidFormer interface {
	Valid(*Validation)
}

func (v *Validation) apply(chk Validator, obj interface{}) *Result {
	if nil == obj {
		if chk.IsSatisfied(obj) {
			return &Result{Ok: true}
		}
	} else if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		if reflect.ValueOf(obj).IsNil() {
			if chk.IsSatisfied(nil) {
				return &Result{Ok: true}
			}
		} else {
			if chk.IsSatisfied(reflect.ValueOf(obj).Elem().Interface()) {
				return &Result{Ok: true}
			}
		}
	} else if chk.IsSatisfied(obj) {
		return &Result{Ok: true}
	}

	// Add the error to the validation context.
	key := chk.GetKey()
	Name := key
	Field := ""
	Label := ""
	parts := strings.Split(key, ".")
	if len(parts) == 3 {
		Field = parts[0]
		Name = parts[1]
		Label = parts[2]
		if len(Label) == 0 {
			Label = Field
		}
	}

	err := &ErrorValidation{
		Message:    Label + chk.DefaultMessage(),
		Key:        key,
		Name:       Name,
		Field:      Field,
		Value:      obj,
		Tmpl:       MessageTmpfs[Name],
		LimitValue: chk.GetLimitValue(),
	}
	v.setError(err)

	// Also return it in the result.
	return &Result{
		Ok:    false,
		Error: err,
	}
}

package valid

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validator interface
type Validator interface {
	IsSatisfied(interface{}) bool
	DefaultMessage() string
	GetKey() string
	GetLimitValue() interface{}
}

// ValidFunc Valid function type
type ValidFunc struct {
	Name   string
	Params []interface{}
}

// CustomFunc is for custom validate function
type CustomFunc func(v *Validation, obj interface{}, key string)

// ReflectFunc Validate function map
type ReflectFunc map[string]reflect.Value

// A Validation context manages data validation and error messages.
type Validation struct {
	// if this field set true, in struct tag valid
	// if the struct field vale is empty
	// it will skip those valid functions, see CanSkipFunc
	RequiredFirst bool

	Errors    []*ErrorValidation
	ErrorsMap map[string][]*ErrorValidation
}

func init() {
	v := &Validation{}
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !unReflectFunc[m.Name] {
			reflectFunc[m.Name] = m.Func
		}
	}
}

// AddCustomFunc Add a custom function to validation
// The name can not be:
//
//	Clear
//	HasErrors
//	ErrorMap
//	Error
//	Check
//	Valid
//	NoMatch
//
// If the name is same with exists function, it will replace the origin valid function
func AddCustomFunc(name string, f CustomFunc) error {
	if unReflectFunc[name] {
		return fmt.Errorf("invalid function name: %s", name)
	}

	reflectFunc[name] = reflect.ValueOf(f)
	return nil
}

// Clear Clean all ValidationError.
func (v *Validation) Clear() {
	v.Errors = []*ErrorValidation{}
	v.ErrorsMap = nil
}

// HasErrors Has ValidationError nor not.
func (v *Validation) HasErrors() bool {
	return len(v.Errors) > 0
}

// ErrorMap Return the errors mapped by key.
// If there are multiple validation errors associated with a single key, the
// first one "wins".  (Typically the first validation will be the more basic).
func (v *Validation) ErrorMap() map[string][]*ErrorValidation {
	return v.ErrorsMap
}

// Error Add an error to the validation context.
func (v *Validation) Error(message string, args ...interface{}) *Result {
	result := (&Result{
		Ok:    false,
		Error: &ErrorValidation{},
	}).Message(message, args...)
	v.Errors = append(v.Errors, result.Error)
	return result
}

// modify the Parameters type to adapt the function input parameters' type
func parseParam(t reflect.Type, s string) (i interface{}, err error) {
	switch t.Kind() {
	case reflect.Int:
		i, err = strconv.Atoi(s)
	case reflect.Int64:
		if wordSize == 32 {
			return nil, ErrInt64On32
		}
		i, err = strconv.ParseInt(s, 10, 64)
	case reflect.Int32:
		var v int64
		v, err = strconv.ParseInt(s, 10, 32)
		if err == nil {
			i = int32(v)
		}
	case reflect.Int16:
		var v int64
		v, err = strconv.ParseInt(s, 10, 16)
		if err == nil {
			i = int16(v)
		}
	case reflect.Int8:
		var v int64
		v, err = strconv.ParseInt(s, 10, 8)
		if err == nil {
			i = int8(v)
		}
	case reflect.String:
		i = s
	case reflect.Ptr:
		if t.Elem().String() != "regexp.Regexp" {
			err = fmt.Errorf("not support %s", t.Elem().String())
			return
		}
		i, err = regexp.Compile(s)
	default:
		err = fmt.Errorf("not support %s", t.Kind().String())
	}
	return
}

func mergeParam(v *Validation, obj interface{}, params []interface{}) []interface{} {
	return append([]interface{}{v, obj}, params...)
}

// Call validate values with named type string
func (f ReflectFunc) Call(name string, params ...interface{}) (result []reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	if _, ok := f[name]; !ok {
		err = fmt.Errorf("%s does not exist", name)
		return
	}
	if len(params) != f[name].Type().NumIn() {
		err = fmt.Errorf("the number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f[name].Call(in)
	return
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func getValidFunc(f reflect.StructField) (vfs []ValidFunc, err error) {
	tag := f.Tag.Get(ValidTag)
	label := f.Tag.Get(LabelTag)
	if len(tag) == 0 {
		return
	}
	if vfs, tag, err = getRegFunc(tag, f.Name); err != nil {
		return
	}
	fs := strings.Split(tag, ";")
	for _, v := range fs {
		var vf ValidFunc
		if len(v) == 0 {
			continue
		}
		vf, err = parseFunc(v, f.Name, label)
		if err != nil {
			return
		}
		vfs = append(vfs, vf)
	}
	return
}

func getValidPFunc(reflectType reflect.Type) (vfs []ValidFunc, err error) {
	if reflectType.Kind() != reflect.Struct {
		return
	}
	validFunc := make([]ValidFunc, 0)
	for i := 0; i < reflectType.NumField(); i++ {
		tagJson := reflectType.Field(i).Tag.Get("json")
		if tagJson == "" {
			vfs, err = getValidPFunc(reflectType.Field(i).Type)
			for _, v := range vfs {
				validFunc = append(validFunc, v)
			}
			continue
		}
		tag := reflectType.Field(i).Tag.Get(ValidTag)
		label := reflectType.Field(i).Tag.Get(LabelTag)
		if len(tag) == 0 {
			vfs = validFunc
			return
		}
		if vfs, tag, err = getRegFunc(tag, reflectType.Field(i).Name); err != nil {
			vfs = validFunc
			return
		}
		fs := strings.Split(tag, ";")
		for _, v := range fs {
			var vf ValidFunc
			if len(v) == 0 {
				continue
			}
			vf, err = parseFunc(v, reflectType.Field(i).Name, label)
			if err != nil {
				vfs = validFunc
				return
			}
			validFunc = append(validFunc, vf)
		}
	}
	vfs = validFunc
	return
}

// Get Match function
// May be Get NoMatch function in the future
func getRegFunc(tag, key string) (vfs []ValidFunc, str string, err error) {
	tag = strings.TrimSpace(tag)
	index := strings.Index(tag, "Match(/")
	if index == -1 {
		str = tag
		return
	}
	end := strings.LastIndex(tag, "/)")
	if end < index {
		err = fmt.Errorf("invalid Match function")
		return
	}
	reg, err := regexp.Compile(tag[index+len("Match(/") : end])
	if err != nil {
		return
	}
	vfs = []ValidFunc{{"Match", []interface{}{reg, key + ".Match"}}}
	str = strings.TrimSpace(tag[:index]) + strings.TrimSpace(tag[end+len("/)"):])
	return
}

func parseFunc(vFunc, key string, label string) (v ValidFunc, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	vFunc = strings.TrimSpace(vFunc)
	start := strings.Index(vFunc, "(")
	var num int

	// doesn't need parameter valid function
	if start == -1 {
		if num, err = numIn(vFunc); err != nil {
			return
		}
		if num != 0 {
			err = fmt.Errorf("%s require %d parameters", vFunc, num)
			return
		}
		v = ValidFunc{vFunc, []interface{}{key + "." + vFunc + "." + label}}
		return
	}

	end := strings.Index(vFunc, ")")
	if end == -1 {
		err = fmt.Errorf("invalid valid function")
		return
	}

	name := strings.TrimSpace(vFunc[:start])
	if num, err = numIn(name); err != nil {
		return
	}

	params := strings.Split(vFunc[start+1:end], ",")
	// the num of param must be equal
	if num != len(params) {
		err = fmt.Errorf("%s require %d parameters", name, num)
		return
	}

	tParams, err := trim(name, key+"."+name+"."+label, params)
	if err != nil {
		return
	}
	v = ValidFunc{name, tParams}
	return
}

func numIn(name string) (num int, err error) {
	fn, ok := reflectFunc[name]
	if !ok {
		err = fmt.Errorf("doesn't exists %s valid function", name)
		return
	}
	// sub *Validation obj and key
	num = fn.Type().NumIn() - 3
	return
}

func trim(name, key string, s []string) (ts []interface{}, err error) {
	ts = make([]interface{}, len(s), len(s)+1)
	fn, ok := reflectFunc[name]
	if !ok {
		err = fmt.Errorf("doesn't exists %s valid function", name)
		return
	}
	for i := 0; i < len(s); i++ {
		var param interface{}
		// skip *Validation and obj params
		if param, err = parseParam(fn.Type().In(i+2), strings.TrimSpace(s[i])); err != nil {
			return
		}
		ts[i] = param
	}
	ts = append(ts, key)
	return
}

// AddError key must like aa.bb.cc or aa.bb.
// AddError adds independent error message for the provided key
func (v *Validation) AddError(key, message string) {
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
		Message: Label + " " + message,
		Key:     key,
		Name:    Name,
		Field:   Field,
	}
	v.setError(err)
}

func (v *Validation) setError(err *ErrorValidation) {
	v.Errors = append(v.Errors, err)
	if v.ErrorsMap == nil {
		v.ErrorsMap = make(map[string][]*ErrorValidation)
	}
	if _, ok := v.ErrorsMap[err.Field]; !ok {
		v.ErrorsMap[err.Field] = []*ErrorValidation{}
	}
	v.ErrorsMap[err.Field] = append(v.ErrorsMap[err.Field], err)
}

// SetError SetCalendar error message for one field in ValidationError
func (v *Validation) SetError(fieldName string, errMsg string) *ErrorValidation {
	err := &ErrorValidation{Key: fieldName, Field: fieldName, Tmpl: errMsg, Message: errMsg}
	v.setError(err)
	return err
}

// Check Apply a group of validators to a field, in order, and return the
// ValidationResult from the first one that fails, or the last one that
// succeeds.
func (v *Validation) Check(obj interface{}, checks ...Validator) *Result {
	var result *Result
	for _, check := range checks {
		result = v.apply(check, obj)
		if !result.Ok {
			return result
		}
	}
	return result
}

//
//// 层级递增解析tag
//func GetReflectTag(reflectType reflect.Type, vfs *[]ValidFunc) {
//	if reflectType.Kind() != reflect.Struct {
//		return
//	}
//	for i := 0; i < reflectType.NumField(); i++ {
//		tag := reflectType.Field(i).Tag.Get("json")
//		if tag == "" {
//			GetReflectTag(reflectType.Field(i).Type, buf)
//			continue
//		}
//		buf.WriteString("`")
//		buf.WriteString(tag)
//		buf.WriteString("`,")
//	}
//}
//
//// 根据model中表模型的json标签获取表字段
//// 将select* 变为对应的字段名
//func GetColSQL(model interface{}) (sql string) {
//	var buf bytes.Buffer
//
//	typ := reflect.TypeOf(model)
//	for i := 0; i < typ.NumField(); i++ {
//		tag := typ.Field(i).Tag.Get("json")
//		if tag == "" {
//			GetReflectTag(typ.Field(i).Type, &buf)
//			continue
//		}
//		// sql += "`" + tag + "`,"
//		buf.WriteString("`")
//		buf.WriteString(tag)
//		buf.WriteString("`,")
//	}
//	sql = string(buf.Bytes()[:buf.Len()-1]) //去掉点,
//	return sql
//}

// Valid Validate a struct.
// the obj parameter must be a struct or a struct pointer
func (v *Validation) Valid(obj interface{}) (b bool, err error) {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)

	switch {
	case isStruct(objT):
	case isStructPtr(objT):
		objT = objT.Elem()
		objV = objV.Elem()
	default:
		err = fmt.Errorf("%v must be a struct or a struct pointer", obj)
		return
	}

	for i := 0; i < objT.NumField(); i++ {
		var vfs []ValidFunc
		tag := objT.Field(i).Tag.Get("json")

		if tag == "" {
			if vfs, err = getValidPFunc(objT.Field(i).Type); err != nil {
				return
			}
		} else {
			if vfs, err = getValidFunc(objT.Field(i)); err != nil {
				return
			}
		}
		//keys := objT.Field(i).Tag.Get("json")
		//fmt.Println(keys)
		var hasRequired bool
		for _, vf := range vfs {
			if vf.Name == "Required" {
				hasRequired = true
			}
			currentField := objV.Field(i).Interface()

			if objV.Field(i).Kind() == reflect.Slice {
				valSlice := reflect.ValueOf(currentField)
				for ii := 0; ii < valSlice.Len(); ii++ {
					v.Valid(valSlice.Index(ii).Interface())
				}
			}

			if objV.Field(i).Kind() == reflect.Ptr {
				v.Valid(currentField)
				//if objV.Field(i).IsNil() {
				//	currentField = ""
				//} else {
				//	currentField = objV.Field(i).Elem().Interface()
				//}
			}

			if objV.Field(i).Kind() == reflect.Struct {
				v.Valid(currentField)
				//if objV.Field(i).IsNil() {
				//	currentField = ""
				//} else {
				//	currentField = objV.Field(i).Elem().Interface()
				//}
			}

			chk := Required{}.IsSatisfied(currentField)
			if !hasRequired && v.RequiredFirst && !chk {
				if _, ok := CanSkipFunc[vf.Name]; ok {
					continue
				}
			}

			if _, err = reflectFunc.Call(vf.Name, mergeParam(v, objV.Field(i).Interface(), vf.Params)...); err != nil {
				return
			}
		}
	}

	if !v.HasErrors() {
		if form, ok := obj.(ValidFormer); ok {
			form.Valid(v)
		}
	}

	return !v.HasErrors(), nil
}

// RecursiveValid Recursively validate a struct.
// Step1: Validate by v.Valid
// Step2: If pass on step1, then reflect obj 's fields
// Step3: Do the Recursively validation to all struct or struct pointer fields
func (v *Validation) RecursiveValid(objc interface{}) (bool, error) {
	//Step 1: validate obj itself firstly
	// fails if objc is not struct
	pass, err := v.Valid(objc)
	if err != nil || !pass {
		return pass, err // Stop recursive validation
	}
	// Step 2: Validate struct s struct fields
	objT := reflect.TypeOf(objc)
	objV := reflect.ValueOf(objc)

	if isStructPtr(objT) {
		objT = objT.Elem()
		objV = objV.Elem()
	}

	for i := 0; i < objT.NumField(); i++ {

		t := objT.Field(i).Type

		// Recursive applies to struct or pointer to structs fields
		if isStruct(t) || isStructPtr(t) {
			// Step 3: do the recursive validation
			// Only valid the Public field recursively
			if objV.Field(i).CanInterface() {
				pass, err = v.RecursiveValid(objV.Field(i).Interface())
			}
		}
	}
	return pass, err
}

func (v *Validation) CanSkipAlso(skipFunc string) {
	if _, ok := CanSkipFunc[skipFunc]; !ok {
		CanSkipFunc[skipFunc] = struct{}{}
	}
}

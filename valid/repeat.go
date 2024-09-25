package valid

import (
	"reflect"
)

// Repeat Requires an integer to be within Min, Max inclusive.
type Repeat struct {
	Key string
}

// IsSatisfied judge whether obj is valid
// not support int64 on 32-bit platform
func (r *Repeat) IsSatisfied(obj interface{}) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Slice {
		return false
	}
	switch val := obj.(type) {
	case []string:
		return len(removeDuplicates(val)) == v.Len()
	case []int:
		return len(removeDuplicates(val)) == v.Len()
	case []uint:
		return len(removeDuplicates(val)) == v.Len()
	case []int8:
		return len(removeDuplicates(val)) == v.Len()
	case []uint8:
		return len(removeDuplicates(val)) == v.Len()
	case []int16:
		return len(removeDuplicates(val)) == v.Len()
	case []uint16:
		return len(removeDuplicates(val)) == v.Len()
	case []uint32:
		return len(removeDuplicates(val)) == v.Len()
	case []int32:
		return len(removeDuplicates(val)) == v.Len()
	case []int64:
		return len(removeDuplicates(val)) == v.Len()
	case []uint64:
		return len(removeDuplicates(val)) == v.Len()
	default:
		return false
	}

	return true
}

// DefaultMessage return the default error message
func (r *Repeat) DefaultMessage() string {
	return MessageTmpfs["Repeat"]
}

// GetKey return the r.Key
func (r *Repeat) GetKey() string {
	return r.Key
}

// GetLimitValue return nil now
func (r *Repeat) GetLimitValue() interface{} {
	return nil
}

// Equatable 接口用于比较切片中的元素是否相等
type Equatable interface {
	comparable
}

// Repeat 泛型函数，去除切片中的重复元素
func removeDuplicates[T Equatable](slice []T) []T {
	var result []T
	seen := make(map[T]struct{}) // 使用空结构体来减少内存占用

	for _, v := range slice {
		if _, found := seen[v]; !found {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

//
//// RemoveDuplicatesStruct 泛型函数，去除切片中的重复结构体元素
//// 需要提供一个自定义的比较函数来比较结构体
//func RemoveDuplicatesStruct[T any, K comparable](slice []T, keyFunc func(T) K) []T {
//	var result []T
//	seen := make(map[K]struct{})
//
//	for _, v := range slice {
//		key := keyFunc(v) // 使用自定义函数提取比较字段
//		if _, found := seen[key]; !found {
//			seen[key] = struct{}{}
//			result = append(result, v)
//		}
//	}
//	return result
//}

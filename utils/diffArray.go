package utils

import (
	"fmt"
	"reflect"
)

type StructUniqueKey interface {
	GetStructUniqueKey() string
}

type Same struct {
	Left  StructUniqueKey
	Right StructUniqueKey
}

func RightDiffArray[T StructUniqueKey](a, b []T) []T {
	m := make(map[string]T)
	for _, v := range a {
		m[v.GetStructUniqueKey()] = v
	}
	var str []T
	for _, v := range b {
		if _, ok := m[v.GetStructUniqueKey()]; !ok {
			str = append(str, v)
		}
	}
	return str
}

type ArrayUnique struct {
	LeftArray     interface{}
	leftArrayKey  map[string]StructUniqueKey
	RightArray    interface{}
	rightArrayKey map[string]StructUniqueKey
}

func (a *ArrayUnique) getArrayObjectKey(leftOrRight bool) {
	var arrayObject interface{}
	if leftOrRight {
		arrayObject = a.LeftArray
		if a.leftArrayKey != nil {
			return
		}
		a.leftArrayKey = make(map[string]StructUniqueKey)
	} else {
		arrayObject = a.RightArray
		if a.rightArrayKey != nil {
			return
		}
		a.rightArrayKey = make(map[string]StructUniqueKey)
	}
	v := reflect.ValueOf(arrayObject)
	if !(v.Kind() == reflect.Slice) {
		panic(fmt.Sprintf("arrayObject must be a slice, but got %v", v.Kind()))
	}

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface().(StructUniqueKey)
		if leftOrRight {
			a.leftArrayKey[item.GetStructUniqueKey()] = item
		} else {
			a.rightArrayKey[item.GetStructUniqueKey()] = item
		}
	}

}

// GetInLeftObject 获取差集,在左侧存在，但是不存在右侧
func (a *ArrayUnique) GetInLeftObject() []StructUniqueKey {
	if a.leftArrayKey == nil {
		a.getArrayObjectKey(true)
	}
	if a.rightArrayKey == nil {
		a.getArrayObjectKey(false)
	}
	var result []StructUniqueKey
	for k, v := range a.rightArrayKey {
		if _, ok := a.leftArrayKey[k]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// GetInRightObject 获取差集,在右侧存在，但是不存在左侧
func (a *ArrayUnique) GetInRightObject() []StructUniqueKey {
	if a.leftArrayKey == nil {
		a.getArrayObjectKey(true)
	}
	if a.rightArrayKey == nil {
		a.getArrayObjectKey(false)
	}
	var result []StructUniqueKey
	for k, v := range a.leftArrayKey {
		if _, ok := a.rightArrayKey[k]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// GetIntersection 获取交集
func (a *ArrayUnique) GetIntersection() []Same {
	if a.leftArrayKey == nil {
		a.getArrayObjectKey(true)
	}
	if a.rightArrayKey == nil {
		a.getArrayObjectKey(false)
	}

	var result []Same
	for k, v := range a.leftArrayKey {
		if _, ok := a.rightArrayKey[k]; ok {
			result = append(result, Same{Left: v, Right: a.rightArrayKey[k]})
		}
	}

	return result
}

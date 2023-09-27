package utils

import "github.com/mattn/go-sqlite3"

func InArray[T string | int64](val T, array []T) bool {
	for _, v := range array {
		if v == val {
			return true
		}
	}
	return false
}

// ArrayDifference 在slice1但是不在slice2的值
func ArrayDifference[T string | int64](slice1, slice2 []T) []T {
	m := make(map[T]T)
	for _, v := range slice1 {
		m[v] = v
	}
	for _, v := range slice2 {
		if _, ok := m[v]; ok {
			delete(m, v)
		}
	}
	var str []T

	for _, s2 := range m {
		str = append(str, s2)
	}
	return str
}

func IsUniqueErr(err error) bool {
	if sqliteErr, ok := err.(sqlite3.Error); ok {
		if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			return true
		}
	}
	return false
}

type DBFileLock struct {
	S string
}
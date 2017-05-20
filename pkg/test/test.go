package test

import (
	"math/rand"
	"reflect"
)

// ASCII a better string implementation for quick checking urls.
type ASCII []string

// Generate allows ASCII to be used within quickcheck scenarios.
func (ASCII) Generate(r *rand.Rand, size int) reflect.Value {
	var (
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		res   []byte
	)

	for i := 0; i < size; i++ {
		pos := r.Intn(len(chars) - 1)
		res = append(res, chars[pos])
	}

	return reflect.ValueOf([]string{string(res)})
}

func (a ASCII) String() string {
	return a[0]
}

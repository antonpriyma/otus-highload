package random

import (
	"math/rand"
	"time"
)

type StringGenerator interface {
	RandString(n uint8) string
}

const (
	Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type stringGenerator struct {
	charset string
	rand    *rand.Rand
}

func NewStringGenerator(charset string) StringGenerator {
	return stringGenerator{
		charset: charset,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())), // nolint:gosec
	}
}

func (r stringGenerator) RandString(length uint8) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = r.charset[r.rand.Int63()%int64(len(r.charset))]
	}
	return string(b)
}

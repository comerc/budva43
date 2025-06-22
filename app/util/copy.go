package util

import (
	"log"

	"github.com/vmihailenco/msgpack/v5"
)

func SimpleCopy[T any](from *T) *T {
	result := *from
	return &result
}

// Copy копирует любую структуру, про ограничения: "doc/COPY.md"
func Copy[T any](from *T) *T {
	var err error
	var b []byte
	b, err = msgpack.Marshal(from)
	if err != nil {
		log.Panic("Copy: ", err)
	}
	to := new(T)
	err = msgpack.Unmarshal(b, to)
	if err != nil {
		log.Panic("Copy: ", err)
	}
	return to
}

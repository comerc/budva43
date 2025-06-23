package util

import (
	"encoding/json"
)

func SimpleCopy[T any](from *T) *T {
	result := *from
	return &result
}

// DeepCopy копирует любую структуру, про ограничения: "doc/COPY.md"
func DeepCopy[T any](from *T) (*T, error) {
	var err error
	var b []byte
	b, err = json.Marshal(from)
	if err != nil {
		return nil, err
	}
	to := new(T)
	err = json.Unmarshal(b, to)
	if err != nil {
		return nil, err
	}
	return to, nil
}

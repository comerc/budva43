package util

import (
	"log"
	"slices"
	"strconv"
	"time"
	"unicode/utf16"

	"github.com/vmihailenco/msgpack/v5"
)

// RuneCountForUTF16 возвращает количество символов в строке, учитывая UTF-16
func RuneCountForUTF16(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// ConvertToInt преобразует строку в целое число
func ConvertToInt[T int | int64](s string) T {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Panic("ConvertToInt: ", err)
	}
	return T(i)
}

// Distinct возвращает уникальные элементы из массива строк
func Distinct(a []string) []string {
	slices.Sort(a)
	return slices.Compact(a)
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

// GetCurrentDate возвращает текущую дату в формате YYYY-MM-DD
func GetCurrentDate() string {
	return time.Now().Format("2000-12-31")
}

// ErrSet представляет собой коллекцию ошибок, которые могут возникнуть при shutdown
type ErrSet struct {
	errors []error
}

// Add добавляет ошибку в набор ошибок, если она не nil
func (e *ErrSet) Add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// GetErrors возвращает набор ошибок
func (e *ErrSet) GetErrors() []error {
	return e.errors
}

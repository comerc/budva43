package util

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"
	"unicode/utf16"
)

// deprecated
// RuneCountForUTF16 возвращает количество символов в строке, учитывая UTF-16
func RuneCountForUTF16(s string) int {
	return len(EncodeToUTF16(s))
}

// EncodeToUTF16 преобразует строку в срез UTF-16
func EncodeToUTF16(s string) []uint16 {
	return utf16.Encode([]rune(s))
}

// DecodeFromUTF16 преобразует срез UTF-16 в строку
func DecodeFromUTF16(utf16s []uint16) string {
	return string(utf16.Decode(utf16s))
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

// GetCurrentDate возвращает текущую дату в формате YYYY-MM-DD
func GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

func NewFuncWithIndex(prefix string) func() string {
	index := -1
	return func() string {
		index++
		return fmt.Sprintf("%s.%d", prefix, index)
	}
}

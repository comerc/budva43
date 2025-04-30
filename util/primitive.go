package util

import (
	"encoding/binary"
	"log"
	"strconv"
	"unicode/utf16"
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

// Uint64ToBytes преобразует uint64 в байтовый массив
func Uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

// BytesToUint64 преобразует байтовый массив в uint64
func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

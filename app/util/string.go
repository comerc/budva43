package util

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

// EncodeToUTF16 преобразует строку в срез UTF-16
func EncodeToUTF16(s string) []uint16 {
	return utf16.Encode([]rune(s))
}

// DecodeFromUTF16 преобразует срез UTF-16 в строку
func DecodeFromUTF16(a []uint16) string {
	return string(utf16.Decode(a))
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

var reMarkdown *regexp.Regexp

func initReMarkdown() {
	s := "_ * ( ) ~ ` > # + = | { } . ! \\[ \\] \\-"
	a := strings.Split(s, " ")
	reMarkdown = regexp.MustCompile("[" + strings.Join(a, "") + "]")
}

// EscapeMarkdown экранирует markdown спецсимволы
func EscapeMarkdown(text string) string {
	// s := "_ * ( ) ~ ` > # + = | { } . ! \\[ \\] \\-"
	// a := strings.Split(s, " ")
	// result := text
	// for _, v := range a {
	// 	result = strings.ReplaceAll(result, v, "\\"+v)
	// }
	// return result
	// re := regexp.MustCompile("[" + strings.Join(a, "|") + "]")
	// return re.ReplaceAllString(text, `\$0`)
	return reMarkdown.ReplaceAllString(text, `\$0`)
}

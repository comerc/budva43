package util

import "unicode/utf16"

// RuneCountForUTF16 возвращает количество символов в строке, учитывая UTF-16
func RuneCountForUTF16(s string) int {
	return len(utf16.Encode([]rune(s)))
}

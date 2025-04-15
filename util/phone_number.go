package util

import "strings"

// maskPhoneNumber маскирует номер телефона, заменяя 5 цифр перед последними двумя
// Например, +7 926 111 22 33 становится +7926*****33
func MaskPhoneNumber(phone string) string {
	// Удаляем возможные пробелы и другие разделители
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	// Кол-во символов для маскирования
	const maskedCount = 5
	// Кол-во видимых символов в конце
	const visibleSuffixCount = 2

	// Проверяем минимальную длину номера
	if len(cleanPhone) <= maskedCount+visibleSuffixCount {
		// Если номер слишком короткий, показываем только последние 2 символа
		if len(cleanPhone) <= visibleSuffixCount {
			return "**" // Слишком короткий номер
		}
		// Маскируем все, кроме последних двух
		visibleSuffix := cleanPhone[len(cleanPhone)-visibleSuffixCount:]
		maskLength := len(cleanPhone) - visibleSuffixCount
		mask := strings.Repeat("*", maskLength)
		return mask + visibleSuffix
	}

	// Видимый префикс (всё, кроме последних 7 символов)
	prefixLength := len(cleanPhone) - maskedCount - visibleSuffixCount
	visiblePrefix := cleanPhone[:prefixLength]

	// Последние 2 символа
	visibleSuffix := cleanPhone[len(cleanPhone)-visibleSuffixCount:]

	// Маскированная часть (5 символов)
	mask := strings.Repeat("*", maskedCount)

	return visiblePrefix + mask + visibleSuffix
}

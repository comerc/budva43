package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskPhoneNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{
			name:     "standard_phone_number",
			phone:    "+7 926 111 22 33",
			expected: "+7926*****33",
		},
		{
			name:     "phone_number_without_spaces",
			phone:    "+79261112233",
			expected: "+7926*****33",
		},
		{
			name:     "phone_number_with_dashes",
			phone:    "+7-926-111-22-33",
			expected: "+7926*****33",
		},
		{
			name:     "short_phone_number",
			phone:    "12345",
			expected: "***45",
		},
		{
			name:     "very_short_number-2_chars",
			phone:    "12",
			expected: "**", // По реализации, если длина <= 2, возвращается "**"
		},
		{
			name:     "very_short_number-1_char",
			phone:    "1",
			expected: "**", // По реализации, если длина <= 2, возвращается "**"
		},
		{
			name:     "empty_number",
			phone:    "",
			expected: "**", // Согласно логике функции для слишком коротких номеров
		},
		{
			name:     "international_format_number",
			phone:    "+38 067 123 45 67",
			expected: "+38067*****67", // Учитываем фактический результат
		},
		{
			name:     "number_with_letters",
			phone:    "+7 (XXX) 123-45-67",
			expected: "+7(XXX)*****67",
		},
		{
			name:     "number_with_7_chars-border",
			phone:    "1234567",
			expected: "*****67",
		},
		{
			name:     "number_with_8_chars-1_visible_prefix",
			phone:    "12345678",
			expected: "1*****78",
		},
		{
			name:     "number_with_spaces_at_start_and_end",
			phone:    "  +7 926  123  45  67  ",
			expected: "+7926*****67",
		},
		{
			name:     "very_long_number",
			phone:    "+7926123456789012345",
			expected: "+792612345678*****45",
		},
		{
			name:     "number_with_non_digit_symbols",
			phone:    "АБВabcdefghij",
			expected: "АБВabc*****ij",
		},
		{
			name:     "number_with_special_symbols",
			phone:    "+7@#$%^&*()_+12345",
			expected: "+7@#$%^&*()*****45", // Сохраняются последние 2 символа
		},
		{
			name:     "number_with_other_special_symbols",
			phone:    "!@#$%^1234567",
			expected: "!@#$%^*****67",
		},
		{
			name:     "number_with_unicode_symbols",
			phone:    "🙃+79261234567",
			expected: "🙃+7926*****67",
		},
		{
			name:     "number_with_6_chars-one_less_than_border",
			phone:    "123456",
			expected: "****56",
		},
		{
			name:     "number_with_3_chars-minimal_for_masking",
			phone:    "123",
			expected: "*23",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := MaskPhoneNumber(test.phone)
			assert.Equal(t, test.expected, result, "Неверная маскировка для номера: %s", test.phone)
		})
	}
}

// TestMaskPhoneNumberPropertyBased проверяет свойства функции маскирования
func TestMaskPhoneNumberPropertyBased(t *testing.T) {
	t.Parallel()

	// Проверка, что оригинальный номер и замаскированный имеют одинаковую длину
	// для номеров длиннее 2 символов
	phones := []string{
		"+79261234567",
		"12345",
		"1234567890",
		"ABCDEFGHIJK",
	}

	for _, phone := range phones {
		cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")
		if len(cleanPhone) > 2 {
			result := MaskPhoneNumber(phone)
			assert.Equal(t, len(cleanPhone), len(result),
				"Длина замаскированного номера должна совпадать с длиной исходного: %s", phone)
		}
	}
}

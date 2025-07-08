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
			name:     "Standard phone number",
			phone:    "+7 926 111 22 33",
			expected: "+7926*****33",
		},
		{
			name:     "Phone number without spaces",
			phone:    "+79261112233",
			expected: "+7926*****33",
		},
		{
			name:     "Phone number with dashes",
			phone:    "+7-926-111-22-33",
			expected: "+7926*****33",
		},
		{
			name:     "Short phone number",
			phone:    "12345",
			expected: "***45",
		},
		{
			name:     "Very short number (2 chars)",
			phone:    "12",
			expected: "**", // По реализации, если длина <= 2, возвращается "**"
		},
		{
			name:     "Very short number (1 char)",
			phone:    "1",
			expected: "**", // По реализации, если длина <= 2, возвращается "**"
		},
		{
			name:     "Empty number",
			phone:    "",
			expected: "**", // Согласно логике функции для слишком коротких номеров
		},
		{
			name:     "International format number",
			phone:    "+38 067 123 45 67",
			expected: "+38067*****67", // Учитываем фактический результат
		},
		{
			name:     "Number with letters",
			phone:    "+7 (XXX) 123-45-67",
			expected: "+7(XXX)*****67",
		},
		{
			name:     "Number with 7 chars (border)",
			phone:    "1234567",
			expected: "*****67",
		},
		{
			name:     "Number with 8 chars (1 visible prefix)",
			phone:    "12345678",
			expected: "1*****78",
		},
		{
			name:     "Number with spaces at start and end",
			phone:    "  +7 926  123  45  67  ",
			expected: "+7926*****67",
		},
		{
			name:     "Very long number",
			phone:    "+7926123456789012345",
			expected: "+792612345678*****45",
		},
		{
			name:     "Number with non-digit symbols",
			phone:    "АБВabcdefghij",
			expected: "АБВabc*****ij",
		},
		{
			name:     "Number with special symbols",
			phone:    "+7@#$%^&*()_+12345",
			expected: "+7@#$%^&*()*****45", // Сохраняются последние 2 символа
		},
		{
			name:     "Number with other special symbols",
			phone:    "!@#$%^1234567",
			expected: "!@#$%^*****67",
		},
		{
			name:     "Number with unicode symbols",
			phone:    "🙃+79261234567",
			expected: "🙃+7926*****67",
		},
		{
			name:     "Number with 6 chars (one less than border)",
			phone:    "123456",
			expected: "****56",
		},
		{
			name:     "Number with 3 chars (minimal for masking)",
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

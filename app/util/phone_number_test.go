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
			name:     "Стандартный номер телефона",
			phone:    "+7 926 111 22 33",
			expected: "+7926*****33",
		},
		{
			name:     "Номер телефона без пробелов",
			phone:    "+79261112233",
			expected: "+7926*****33",
		},
		{
			name:     "Номер телефона с дефисами",
			phone:    "+7-926-111-22-33",
			expected: "+7926*****33",
		},
		{
			name:     "Короткий номер телефона",
			phone:    "12345",
			expected: "***45",
		},
		{
			name:     "Очень короткий номер (2 символа)",
			phone:    "12",
			expected: "**", // По реализации, если длина <= 2, возвращается "**"
		},
		{
			name:     "Очень короткий номер (1 символ)",
			phone:    "1",
			expected: "**", // По реализации, если длина <= 2, возвращается "**"
		},
		{
			name:     "Пустой номер",
			phone:    "",
			expected: "**", // Согласно логике функции для слишком коротких номеров
		},
		{
			name:     "Номер с международным форматом",
			phone:    "+38 067 123 45 67",
			expected: "+38067*****67", // Учитываем фактический результат
		},
		{
			name:     "Номер с буквами",
			phone:    "+7 (XXX) 123-45-67",
			expected: "+7(XXX)*****67",
		},
		{
			name:     "Номер с 7 символами (граница)",
			phone:    "1234567",
			expected: "*****67",
		},
		{
			name:     "Номер с 8 символами (видимый префикс - 1 символ)",
			phone:    "12345678",
			expected: "1*****78",
		},
		{
			name:     "Номер с пробелами в начале и конце строки",
			phone:    "  +7 926  123  45  67  ",
			expected: "+7926*****67",
		},
		{
			name:     "Очень длинный номер",
			phone:    "+7926123456789012345",
			expected: "+792612345678*****45",
		},
		{
			name:     "Номер, содержащий нецифровые символы",
			phone:    "АБВabcdefghij",
			expected: "АБВabc*****ij",
		},
		{
			name:     "Номер со специальными символами",
			phone:    "+7@#$%^&*()_+12345",
			expected: "+7@#$%^&*()*****45", // Сохраняются последние 2 символа
		},
		{
			name:     "Номер с другими специальными символами",
			phone:    "!@#$%^1234567",
			expected: "!@#$%^*****67",
		},
		{
			name:     "Номер с юникод-символами",
			phone:    "🙃+79261234567",
			expected: "🙃+7926*****67",
		},
		{
			name:     "Номер из 6 символов (на 1 меньше границы)",
			phone:    "123456",
			expected: "****56",
		},
		{
			name:     "Номер из 3 символов (минимальный для маскирования)",
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

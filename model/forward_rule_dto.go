package model

/*
ПРИМЕЧАНИЕ:
Этот файл содержит примеры DTO для правил пересылки.
Данный пример демонстрирует концепцию DTO для преобразования данных между API и внутренними сущностями.
В реальном приложении необходимо адаптировать код в соответствии с актуальной структурой сущностей.
*/

import (
	"errors"

	"github.com/comerc/budva43/entity"
)

// ForwardRuleDTO представляет собой объект передачи данных для правила пересылки
// Используется для API создания и обновления правил пересылки
type ForwardRuleDTO struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	From        int64   `json:"from"` // ID чата источника
	To          []int64 `json:"to"`   // ID чатов назначения
	Active      bool    `json:"active"`
	SendCopy    bool    `json:"send_copy"`    // Отправить копию вместо пересылки
	RemoveLinks bool    `json:"remove_links"` // Удалять ссылки
	CopyOnce    bool    `json:"copy_once"`    // Копировать один раз
	Indelible   bool    `json:"indelible"`    // Неудаляемое правило
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
	LastRunAt   string  `json:"last_run_at,omitempty"`

	// Фильтры
	Include   []string `json:"include,omitempty"` // Подстроки для включения сообщений
	Exclude   []string `json:"exclude,omitempty"` // Подстроки для исключения сообщений
	MediaOnly bool     `json:"media_only"`        // Только медиа сообщения
	TextOnly  bool     `json:"text_only"`         // Только текстовые сообщения

	// Настройки замены
	Replacements map[string]string `json:"replacements,omitempty"` // Замены текста (ключ: оригинал, значение: замена)
}

// ForwardRuleBatchDTO представляет список правил пересылки для API
type ForwardRuleBatchDTO struct {
	Rules []*ForwardRuleDTO `json:"rules"`
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

/*
ПРИМЕЧАНИЕ:
Следующие методы являются ШАБЛОНАМИ (stub methods), которые демонстрируют идею
преобразования между DTO и сущностями. В реальном коде эти методы должны быть
реализованы в соответствии с актуальной структурой сущностей.
*/

// Пример метода для демонстрации концепции преобразования сущности в DTO
// ЭТО ТОЛЬКО ПРИМЕР - не для реального использования
func ExampleForwardRuleToDTO(rule *entity.ForwardRule) *ForwardRuleDTO {
	// В реальном коде здесь было бы преобразование из entity.ForwardRule в ForwardRuleDTO
	// с использованием фактических полей entity.ForwardRule

	// Просто возвращаем заглушку для примера
	return &ForwardRuleDTO{
		ID:       "example-id",
		Name:     "Example Rule",
		From:     123456789,
		To:       []int64{987654321},
		Active:   true,
		SendCopy: true,
	}
}

// Пример метода для демонстрации концепции преобразования DTO в сущность
// ЭТО ТОЛЬКО ПРИМЕР - не для реального использования
func ExampleDTOToForwardRule(dto *ForwardRuleDTO) (*entity.ForwardRule, error) {
	if dto == nil {
		return nil, errors.New("forward rule DTO is nil")
	}

	// Пример базовой валидации
	if dto.From == 0 {
		return nil, errors.New("source chat ID is required")
	}

	if len(dto.To) == 0 {
		return nil, errors.New("at least one destination chat ID is required")
	}

	// В реальном коде здесь было бы создание entity.ForwardRule
	// с заполнением его полей из dto

	// Возвращаем заглушку, так как мы не знаем точную структуру entity.ForwardRule
	return &entity.ForwardRule{}, nil
}

// Пример метода для демонстрации создания коллекции DTO из коллекции сущностей
// ЭТО ТОЛЬКО ПРИМЕР - не для реального использования
func ExampleCreateForwardRuleBatchDTO(rules []*entity.ForwardRule, page, size, total int) *ForwardRuleBatchDTO {
	// В реальном коде здесь было бы преобразование каждой сущности в DTO

	// Просто создаем примерную коллекцию для демонстрации
	dtos := make([]*ForwardRuleDTO, 0, 2)
	dtos = append(dtos, &ForwardRuleDTO{
		ID:   "example-1",
		Name: "Example Rule 1",
		From: 111111111,
		To:   []int64{222222222},
	})
	dtos = append(dtos, &ForwardRuleDTO{
		ID:   "example-2",
		Name: "Example Rule 2",
		From: 333333333,
		To:   []int64{444444444},
	})

	return &ForwardRuleBatchDTO{
		Rules: dtos,
		Total: total,
		Page:  page,
		Size:  size,
	}
}

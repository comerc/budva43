package monitoring

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// MetricType тип метрики
type MetricType string

const (
	// MetricTypeCounter счетчик (только возрастает)
	MetricTypeCounter MetricType = "counter"
	// MetricTypeGauge значение, которое может меняться в обе стороны
	MetricTypeGauge MetricType = "gauge"
	// MetricTypeHistogram гистограмма для измерения распределения значений
	MetricTypeHistogram MetricType = "histogram"
)

// MetricValue значение метрики
type MetricValue struct {
	Value     float64
	Timestamp time.Time
}

// Metric структура метрики
type Metric struct {
	Name        string
	Type        MetricType
	Description string
	Values      []MetricValue
	Sum         float64   // для гистограмм и счетчиков
	Min         float64   // для гистограмм
	Max         float64   // для гистограмм
	Count       int       // количество измерений
	LastUpdate  time.Time // время последнего обновления
}

// Service предоставляет методы для мониторинга и сбора метрик
type Service struct {
	log *slog.Logger
	//
	metrics        map[string]*Metric
	metricsHistory map[string][]MetricValue
	historyLimit   int
	startTime      time.Time
	mutex          sync.RWMutex
}

// New создает новый экземпляр сервиса мониторинга
func New(historyLimit int) *Service {
	if historyLimit <= 0 {
		historyLimit = 1000 // значение по умолчанию
	}

	return &Service{
		log: slog.With("module", "service.monitoring"),
		//
		metrics:        make(map[string]*Metric),
		metricsHistory: make(map[string][]MetricValue),
		historyLimit:   historyLimit,
		startTime:      time.Now(),
	}
}

// RegisterMetric регистрирует новую метрику
func (s *Service) RegisterMetric(name string, metricType MetricType, description string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.metrics[name]; exists {
		return fmt.Errorf("метрика с именем %s уже существует", name)
	}

	metric := &Metric{
		Name:        name,
		Type:        metricType,
		Description: description,
		Values:      make([]MetricValue, 0),
		LastUpdate:  time.Now(),
	}

	// Инициализация начальных значений в зависимости от типа
	switch metricType {
	case MetricTypeCounter:
		metric.Sum = 0
	case MetricTypeGauge:
		// Для gauge не требуется инициализация
	case MetricTypeHistogram:
		metric.Sum = 0
		metric.Min = 0
		metric.Max = 0
		metric.Count = 0
	}

	s.metrics[name] = metric
	s.metricsHistory[name] = make([]MetricValue, 0, s.historyLimit)

	return nil
}

// IncrementCounter увеличивает счетчик на указанное значение
func (s *Service) IncrementCounter(name string, value float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metric, exists := s.metrics[name]
	if !exists {
		return fmt.Errorf("метрика %s не найдена", name)
	}

	if metric.Type != MetricTypeCounter {
		return fmt.Errorf("метрика %s не является счетчиком", name)
	}

	// Увеличиваем значение счетчика
	metric.Sum += value
	metric.Count++
	metric.LastUpdate = time.Now()

	// Сохраняем историческое значение
	metricValue := MetricValue{
		Value:     metric.Sum,
		Timestamp: metric.LastUpdate,
	}

	s.addToHistory(name, metricValue)

	return nil
}

// SetGauge устанавливает значение для метрики типа gauge
func (s *Service) SetGauge(name string, value float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metric, exists := s.metrics[name]
	if !exists {
		return fmt.Errorf("метрика %s не найдена", name)
	}

	if metric.Type != MetricTypeGauge {
		return fmt.Errorf("метрика %s не является gauge", name)
	}

	// Устанавливаем значение
	metric.Sum = value
	metric.Count++
	metric.LastUpdate = time.Now()

	// Сохраняем историческое значение
	metricValue := MetricValue{
		Value:     value,
		Timestamp: metric.LastUpdate,
	}

	s.addToHistory(name, metricValue)

	return nil
}

// ObserveHistogram добавляет наблюдение в гистограмму
func (s *Service) ObserveHistogram(name string, value float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metric, exists := s.metrics[name]
	if !exists {
		return fmt.Errorf("метрика %s не найдена", name)
	}

	if metric.Type != MetricTypeHistogram {
		return fmt.Errorf("метрика %s не является гистограммой", name)
	}

	// Обновляем значения гистограммы
	metric.Sum += value
	metric.Count++

	// Обновляем min/max значения
	if metric.Count == 1 {
		// Первое значение
		metric.Min = value
		metric.Max = value
	} else {
		// Обновляем min/max
		if value < metric.Min {
			metric.Min = value
		}
		if value > metric.Max {
			metric.Max = value
		}
	}

	metric.LastUpdate = time.Now()

	// Сохраняем историческое значение
	metricValue := MetricValue{
		Value:     value,
		Timestamp: metric.LastUpdate,
	}

	s.addToHistory(name, metricValue)

	return nil
}

// GetMetric возвращает текущее состояние метрики
func (s *Service) GetMetric(name string) (*Metric, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	metric, exists := s.metrics[name]
	if !exists {
		return nil, fmt.Errorf("метрика %s не найдена", name)
	}

	// Создаем копию метрики для безопасного возврата
	result := *metric
	result.Values = make([]MetricValue, len(metric.Values))
	copy(result.Values, metric.Values)

	return &result, nil
}

// GetAllMetrics возвращает все метрики
func (s *Service) GetAllMetrics() map[string]*Metric {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]*Metric, len(s.metrics))

	// Копируем все метрики в результат
	for name, metric := range s.metrics {
		metricCopy := *metric
		metricCopy.Values = make([]MetricValue, len(metric.Values))
		copy(metricCopy.Values, metric.Values)

		result[name] = &metricCopy
	}

	return result
}

// GetMetricHistory возвращает историю значений метрики
func (s *Service) GetMetricHistory(name string, limit int) ([]MetricValue, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	values, exists := s.metricsHistory[name]
	if !exists {
		return nil, fmt.Errorf("метрика %s не найдена", name)
	}

	// Если лимит не указан или больше количества значений, берем все значения
	if limit <= 0 || limit > len(values) {
		limit = len(values)
	}

	// Копируем последние limit значений
	result := make([]MetricValue, limit)
	startIndex := len(values) - limit
	copy(result, values[startIndex:])

	return result, nil
}

// ResetMetric сбрасывает значение метрики к начальному состоянию
func (s *Service) ResetMetric(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metric, exists := s.metrics[name]
	if !exists {
		return fmt.Errorf("метрика %s не найдена", name)
	}

	// Сбрасываем значения в зависимости от типа
	switch metric.Type {
	case MetricTypeCounter:
		metric.Sum = 0
	case MetricTypeGauge:
		metric.Sum = 0
	case MetricTypeHistogram:
		metric.Sum = 0
		metric.Min = 0
		metric.Max = 0
	}

	metric.Count = 0
	metric.LastUpdate = time.Now()

	// Очищаем историю
	s.metricsHistory[name] = make([]MetricValue, 0, s.historyLimit)

	return nil
}

// ResetAllMetrics сбрасывает все метрики к начальному состоянию
func (s *Service) ResetAllMetrics() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Проходим по всем метрикам и сбрасываем их
	for _, metric := range s.metrics {
		switch metric.Type {
		case MetricTypeCounter:
			metric.Sum = 0
		case MetricTypeGauge:
			metric.Sum = 0
		case MetricTypeHistogram:
			metric.Sum = 0
			metric.Min = 0
			metric.Max = 0
		}
		metric.Count = 0
		metric.LastUpdate = time.Now()
	}

	// Очищаем всю историю
	for name := range s.metricsHistory {
		s.metricsHistory[name] = make([]MetricValue, 0, s.historyLimit)
	}
}

// GetUptime возвращает время работы сервиса
func (s *Service) GetUptime() time.Duration {
	return time.Since(s.startTime)
}

// addToHistory добавляет значение в историю метрики
func (s *Service) addToHistory(name string, value MetricValue) {
	history, exists := s.metricsHistory[name]
	if !exists {
		history = make([]MetricValue, 0, s.historyLimit)
		s.metricsHistory[name] = history
	}

	// Добавляем новое значение
	s.metricsHistory[name] = append(s.metricsHistory[name], value)

	// Если превышен лимит, удаляем самое старое значение
	if len(s.metricsHistory[name]) > s.historyLimit {
		s.metricsHistory[name] = s.metricsHistory[name][1:]
	}
}

package util

// ErrSet представляет собой коллекцию ошибок, которые могут возникнуть при shutdown
type ErrSet struct {
	errors []error
}

// Add добавляет ошибку в набор ошибок, если она не nil
func (e *ErrSet) Add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// GetErrors возвращает набор ошибок
func (e *ErrSet) GetErrors() []error {
	return e.errors
}

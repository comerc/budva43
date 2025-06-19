package util

import "sync"

var once sync.Once

// init() - это зло https://habr.com/ru/articles/771858/
// но подходит для реализации синглтона
func init() {
	once.Do(func() {
		ProjectRoot = findProjectRoot()
	})
}

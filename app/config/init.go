package config

import "sync"

var once sync.Once

// init() - это зло https://habr.com/ru/articles/771858/
func Init() {
	once.Do(func() {
		projectRoot = findProjectRoot()
		*cfg = *load()
		MakeDirs()
	})
}

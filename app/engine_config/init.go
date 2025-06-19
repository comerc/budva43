package engine_config

import (
	"log"
	"sync"

	"github.com/comerc/budva43/app/util"
)

var once sync.Once

func init() {
	once.Do(func() {
		initEngineViper(util.ProjectRoot)
		if err := Reload(); err != nil {
			log.Panic(err)
		}
	})
}

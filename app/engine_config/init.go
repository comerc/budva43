package engine_config

import (
	"errors"
	"log"
	"sync"

	"github.com/comerc/budva43/app/util"
)

var once sync.Once

func init() {
	once.Do(func() {
		initEngineViper(util.ProjectRoot)
		if err := Reload(); err != nil {
			var emptySources *ErrEmptySources
			if !errors.As(err, &emptySources) {
				log.Panic(err)
			}
		}
	})
}

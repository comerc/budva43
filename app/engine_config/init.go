package engine_config

import (
	"sync"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/util"
)

var once sync.Once

func init() {
	once.Do(func() {
		initEngineViper(util.ProjectRoot)
		config.Engine = &domain.EngineConfig{}
		Initialize(config.Engine)
	})
}

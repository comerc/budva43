package test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/testing/spylog"
	"github.com/comerc/budva43/app/util"
)

// из-за циклической зависимости app/spylog vs app/log - тесты вынесены в test/log_test.go

type SomeObject struct {
	log *log.Logger
}

func (s *SomeObject) SomeMethod() {
	s.NestedMethod()
}

func (s *SomeObject) NestedMethod() {
	var err error
	defer s.log.ErrorOrDebug(&err, "", "arg2", "val2")

	err = log.NewError("error message", "arg0", "val0")
	err = log.WrapError(err, "arg1", "val1")
}

func TestLog_SomeMethod(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	var err error

	var copy *entity.LogSource
	copy, err = util.DeepCopy(config.LogSource)
	require.NoError(t, err)
	t.Cleanup(func() {
		config.LogSource = copy
	})
	config.LogSource.RelativePath = true

	var o *SomeObject
	spylogHandler := spylog.GetHandler(t.Name(), func() {
		o = &SomeObject{
			log: log.NewLogger(),
		}
	})
	o.SomeMethod()

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)
	record := records[0]

	assert.Equal(t, "error message", record.Message)
	assert.Equal(t, "val0", spylog.GetAttrValue(record, "arg0"))
	assert.Equal(t, "val1", spylog.GetAttrValue(record, "arg1"))
	assert.Equal(t, "val2", spylog.GetAttrValue(record, "arg2"))
	snaps.MatchSnapshot(t, spylog.GetAttrValue(record, "source"))
}

func TestLog_UnwrappedError(t *testing.T) {
	t.Parallel()

	var err error

	var logger *log.Logger

	spylogHandler := spylog.GetHandler(t.Name(), func() {
		logger = log.NewLogger()
	})

	type SomeError struct {
		error
	}

	err = &SomeError{
		error: errors.New("unwrapped error"),
	}

	logger.ErrorOrDebug(&err, "")

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)
	record := records[0]

	assert.Equal(t, "unwrapped error", record.Message)
	assert.Equal(t, "test.SomeError", spylog.GetAttrValue(record, "type"))
}

func TestLog_WrappedError(t *testing.T) {
	t.Parallel()

	var err error

	var logger *log.Logger

	spylogHandler := spylog.GetHandler(t.Name(), func() {
		logger = log.NewLogger()
	})

	type SomeError struct {
		error
	}

	err = &SomeError{
		error: errors.New("wrapped error"),
	}

	err = log.WrapError(err) // !! обёртка

	logger.ErrorOrDebug(&err, "")

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)
	record := records[0]

	assert.Equal(t, "wrapped error", record.Message)
	assert.Equal(t, "test.SomeError", spylog.GetAttrValue(record, "type"))
}

func TestLog_WithPtr(t *testing.T) {
	t.Parallel()

	var logger *log.Logger
	spylogHandler := spylog.GetHandler(t.Name(), func() {
		logger = log.NewLogger()
	})

	func() {
		var (
			a []string
			m map[string]string
			s string
			i int
			f float64
			b bool
			d time.Duration
			t time.Time
			p *int
		)
		var err error
		err = log.NewError("")
		defer logger.ErrorOrDebug(&err, "",
			"a", &a, "m", &m, "s", &s, "i", &i, "f", &f, "b", &b, "d", &d, "t", &t, "p", &p)
		a = []string{"1", "2", "3"}
		m = map[string]string{"a": "1", "b": "2"}
		s = "val"
		i = 123
		f = 1.1
		b = true
		d = 61 * time.Second
		t = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
		p = &i
	}()

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)
	record := records[0]

	assert.Equal(t, "[1 2 3]", spylog.GetAttrValue(record, "a"))
	assert.Equal(t, "map[a:1 b:2]", spylog.GetAttrValue(record, "m"))
	assert.Equal(t, "val", spylog.GetAttrValue(record, "s"))
	assert.Equal(t, "123", spylog.GetAttrValue(record, "i"))
	assert.Equal(t, "1.1", spylog.GetAttrValue(record, "f"))
	assert.Equal(t, "true", spylog.GetAttrValue(record, "b"))
	assert.Equal(t, "1m1s", spylog.GetAttrValue(record, "d"))
	assert.Equal(t, "2025-01-02 03:04:05 +0000 UTC", spylog.GetAttrValue(record, "t"))
	assert.Equal(t, "0x", spylog.GetAttrValue(record, "p")[0:2])
}

func TestLoggerName(t *testing.T) {
	t.Parallel()

	var test = func(t *testing.T) {
		t.Parallel()

		var spylogHandler *spylog.Handler
		var logger *log.Logger
		var loggerName string

		setup := func() {
			logger = log.NewLogger()
			loggerName = log.GetLoggerName()
		}

		run := func() {
			expectedMessage := t.Name()
			err := errors.New(expectedMessage)
			logger.ErrorOrInfo(&err, "")

			records := spylogHandler.GetRecords()
			require.Equal(t, 1, len(records))
			record := records[0]
			assert.Equal(t, expectedMessage, record.Message)
		}

		// Вызываем из разных мест в одной горутине
		spylogHandler = spylog.GetHandler(t.Name(), setup)
		loggerName0 := loggerName
		run()

		// Вызываем из разных мест в одной горутине
		spylogHandler = spylog.GetHandler(t.Name(), setup)
		loggerName1 := loggerName
		run()

		// Проверяем, что loggerName разный
		assert.NotEqual(t, loggerName0, loggerName1)
	}

	for i := range 10 {
		// Запускаем параллельно 10 горутин
		t.Run(fmt.Sprintf("test %d", i), test)
	}
}

package log

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spylog "github.com/comerc/budva43/app/spylog/core" // !! app/spylog/core
	"github.com/comerc/budva43/app/util"
)

//
// dummy comment - для соблюдения номеров строк в тестах
//

func TestMain(m *testing.M) {
	// Исключение: используется без app/spylog - циклическая зависимость
	spylog.CreateHandler(slog.Default()) // init() app/log before slog.Default()
	os.Exit(m.Run())
}

type SomeObject struct {
	log *Logger
}

func (s *SomeObject) SomeMethod() {
	s.NestedMethod()
}

func (s *SomeObject) NestedMethod() {
	var err error
	defer s.log.ErrorOrDebug(&err, "message", "arg2", "val2")

	err = NewError("error", "arg0", "val0")
	err = WrapError(err, "arg1", "val1")
}

func TestSomeMethod(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных опций

	optionsCopy, err := util.DeepCopy(options)
	require.NoError(t, err)
	t.Cleanup(func() {
		options = optionsCopy
	})

	tests := []struct {
		name           string
		sourceType     SourceType
		relativePath   bool
		expectedSource string
	}{
		{
			name:           "simple with absolute path",
			sourceType:     TypeSourceOne,
			relativePath:   false,
			expectedSource: filepath.Join(util.ProjectRoot, "app/log/log_test.go:40 log.(*SomeObject).NestedMethod"),
		},
		{
			name:           "one",
			sourceType:     TypeSourceOne,
			relativePath:   true,
			expectedSource: "app/log/log_test.go:40 log.(*SomeObject).NestedMethod",
		},
		{
			name:           "more",
			sourceType:     TypeSourceMore,
			relativePath:   true,
			expectedSource: "[0=app/log/log_test.go:40 log.(*SomeObject).NestedMethod 1=app/log/log_test.go:33 log.(*SomeObject).SomeMethod 2=app/log/log_test.go:90 log.TestSomeMethod.func2]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options.ErrorSource.Type = test.sourceType
			options.ErrorSource.RelativePath = test.relativePath

			var o *SomeObject
			spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
				o = &SomeObject{
					log: NewLogger("module_name"),
				}
			})
			o.SomeMethod()

			records := spylogHandler.GetRecords()
			require.Equal(t, len(records), 1)

			assert.Equal(t, slog.LevelError, records[0].Level)
			assert.Equal(t, "error", records[0].Message)
			assert.Equal(t, "val0", spylog.GetAttrValue(records[0], "arg0"))
			assert.Equal(t, "val1", spylog.GetAttrValue(records[0], "arg1"))
			assert.Equal(t, "val2", spylog.GetAttrValue(records[0], "arg2"))
			assert.Equal(t, test.expectedSource, spylog.GetAttrValue(records[0], "source"))
		})
	}

}

type SomeError struct {
	error
}

func TestUnwrappedError(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных опций

	{
		optionsCopy, err := util.DeepCopy(options)
		require.NoError(t, err)
		t.Cleanup(func() {
			options = optionsCopy
		})

		options.ErrorSource.Type = TypeSourceOne
		options.ErrorSource.RelativePath = true
	}

	type OtherObject struct {
		log *Logger
	}

	var err error

	var o *OtherObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &OtherObject{
			log: NewLogger("module_name"),
		}
	})

	err = &SomeError{
		error: errors.New("unwrapped error"),
	}
	o.log.ErrorOrInfo(&err, "message", "arg", "val")

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)

	assert.Equal(t, slog.LevelError, records[0].Level)
	assert.Equal(t, "unwrapped error", records[0].Message)
	assert.Equal(t, "val", spylog.GetAttrValue(records[0], "arg"))
	assert.Equal(t, "log.SomeError", spylog.GetAttrValue(records[0], "type"))
	assert.Equal(t, "app/log/log_test.go:140 log.TestUnwrappedError",
		spylog.GetAttrValue(records[0], "source"))
}

func TestLogWithPtr(t *testing.T) {
	t.Parallel()

	type object struct {
		log *Logger
	}

	var o *object
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &object{
			log: NewLogger("module_name"),
		}
	})

	func() {
		var a []string
		var m map[string]string
		var s string
		var i int
		var f float64
		var b bool
		var d time.Duration
		var t time.Time
		var p *int
		defer o.log.ErrorOrDebug(nil, "message",
			"a", &a, "m", &m, "s", &s, "i", &i, "f", &f, "b", &b, "d", &d, "t", &t, "p", &p)
		a = []string{"1", "2", "3"}
		m = map[string]string{"a": "1", "b": "2"}
		s = "321"
		i = 123
		f = 123.456
		b = true
		d = 61 * time.Second
		t = time.Date(2025, 1, 1, 1, 1, 1, 0, time.UTC)
		p = &i
	}()

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)

	assert.Equal(t, slog.LevelDebug, records[0].Level)
	assert.Equal(t, "message", records[0].Message)
	assert.Equal(t, "[1 2 3]", spylog.GetAttrValue(records[0], "a"))
	assert.Equal(t, "map[a:1 b:2]", spylog.GetAttrValue(records[0], "m"))
	assert.Equal(t, "321", spylog.GetAttrValue(records[0], "s"))
	assert.Equal(t, "123", spylog.GetAttrValue(records[0], "i"))
	assert.Equal(t, "123.456", spylog.GetAttrValue(records[0], "f"))
	assert.Equal(t, "true", spylog.GetAttrValue(records[0], "b"))
	assert.Equal(t, "1m1s", spylog.GetAttrValue(records[0], "d"))
	assert.Equal(t, "2025-01-01 01:01:01 +0000 UTC", spylog.GetAttrValue(records[0], "t"))
	assert.Equal(t, "0x", spylog.GetAttrValue(records[0], "p")[0:2])
}

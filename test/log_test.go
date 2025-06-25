package test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/testing/spylog"
	"github.com/comerc/budva43/app/util"
)

// из-за циклической зависимости app/spylog vs app/log - тесты вынесены в test/log_test.go

// dummy comment - для соблюдения номеров строк в тестах
//
//
//
//
//

type SomeObject struct {
	log *log.Logger
}

func (s *SomeObject) SomeMethod() {
	s.NestedMethod()
}

func (s *SomeObject) NestedMethod() {
	var err error
	defer s.log.ErrorOrDebug(&err, "message", "arg2", "val2")

	err = log.NewError("error", "arg0", "val0")
	err = log.WrapError(err, "arg1", "val1")
}

func TestLog_SomeMethod(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	copy, err := util.DeepCopy(config.ErrorSource)
	require.NoError(t, err)
	t.Cleanup(func() {
		config.ErrorSource = copy
	})

	tests := []struct {
		name           string
		sourceType     entity.ErrorSourceType
		relativePath   bool
		expectedSource string
	}{
		{
			name:           "simple with absolute path",
			sourceType:     entity.TypeErrorSourceOne,
			relativePath:   false,
			expectedSource: filepath.Join(util.ProjectRoot, "test/log_test.go:40 test.(*SomeObject).NestedMethod"),
		},
		{
			name:           "one",
			sourceType:     entity.TypeErrorSourceOne,
			relativePath:   true,
			expectedSource: "test/log_test.go:40 test.(*SomeObject).NestedMethod",
		},
		{
			name:           "more",
			sourceType:     entity.TypeErrorSourceMore,
			relativePath:   true,
			expectedSource: "[0=test/log_test.go:40 test.(*SomeObject).NestedMethod 1=test/log_test.go:33 test.(*SomeObject).SomeMethod 2=test/log_test.go:90 test.TestLog_SomeMethod.func2]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config.ErrorSource.Type = test.sourceType
			config.ErrorSource.RelativePath = test.relativePath

			var o *SomeObject
			spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
				o = &SomeObject{
					log: log.NewLogger("module_name"),
				}
			})
			o.SomeMethod()

			records := spylogHandler.GetRecords()
			require.Equal(t, len(records), 1)
			record := records[0]

			assert.Equal(t, "error", record.Message)
			assert.Equal(t, "val0", spylog.GetAttrValue(record, "arg0"))
			assert.Equal(t, "val1", spylog.GetAttrValue(record, "arg1"))
			assert.Equal(t, "val2", spylog.GetAttrValue(record, "arg2"))
			assert.Equal(t, test.expectedSource, spylog.GetAttrValue(record, "source"))
		})
	}

}

type SomeError struct {
	error
}

func TestLog_UnwrappedError(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	{
		copy, err := util.DeepCopy(config.ErrorSource)
		require.NoError(t, err)
		t.Cleanup(func() {
			config.ErrorSource = copy
		})

		config.ErrorSource.Type = entity.TypeErrorSourceOne
		config.ErrorSource.RelativePath = true
	}

	type OtherObject struct {
		log *log.Logger
	}

	var err error

	var o *OtherObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &OtherObject{
			log: log.NewLogger("module_name"),
		}
	})

	err = &SomeError{
		error: errors.New("unwrapped error"),
	}
	o.log.ErrorOrInfo(&err, "message", "arg", "val")

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)
	record := records[0]

	assert.Equal(t, "unwrapped error", record.Message)
	assert.Equal(t, "val", spylog.GetAttrValue(record, "arg"))
	assert.Equal(t, "test.SomeError", spylog.GetAttrValue(record, "type"))
	assert.Equal(t, "test/log_test.go:140 test.TestLog_UnwrappedError",
		spylog.GetAttrValue(record, "source"))
}

func TestLog_WithPtr(t *testing.T) {
	t.Parallel()

	type object struct {
		log *log.Logger
	}

	var o *object
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &object{
			log: log.NewLogger("module_name"),
		}
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
		err = log.NewError("dummy")
		defer o.log.ErrorOrDebug(&err, "dummy",
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

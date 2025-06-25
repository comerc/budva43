package storage

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/dgraph-io/badger/v4"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
)

// Repo определяет интерфейс для работы с хранилищем BadgerDB
type Repo struct {
	log *log.Logger
	//
	db *badger.DB
}

// New создает новый экземпляр репозитория для BadgerDB
func New() *Repo {
	return &Repo{
		log: log.NewLogger("repo.storage"),
		//
		db: nil,
	}
}

// Start устанавливает соединение с базой данных
func (r *Repo) Start(ctx context.Context) error {
	var err error

	opts := badger.DefaultOptions(config.Storage.DatabaseDirectory)
	opts.Logger = NewLogger()
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	r.db = db

	go r.runGarbageCollection(ctx)

	return nil
}

// Close закрывает соединение с базой данных
func (r *Repo) Close() error {
	if r.db != nil {
		err := r.db.Close()
		if err != nil {
			return log.WrapError(err) // внешняя ошибка
		}
	}
	return nil
}

// runGarbageCollection выполняет сборку мусора для базы данных
func (r *Repo) runGarbageCollection(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Агрессивная сборка мусора - продолжаем пока есть что чистить
			for {
				err := r.db.RunValueLogGC(0.7)
				if err == nil {
					// Успешно очистили, пытаемся еще раз
					continue
				}
				if err == badger.ErrNoRewrite {
					// Нет файлов для перезаписи - это нормально, выходим
					err = nil
				}
				// Серьезная ошибка (ErrRejected, закрытая БД и т.д.)
				r.log.ErrorOrDebug(&err, "runGarbageCollection")
				break // Выходим из внутреннего цикла, ждем следующего тика
			}
		}
	}
}

// Increment увеличивает значение по ключу на 1
func (r *Repo) Increment(key string) (uint64, error) {
	var (
		err    error
		result uint64
	)
	// Merge function to add two uint64 numbers
	add := func(existing, _new []byte) []byte {
		return convertUint64ToBytes(ConvertBytesToUint64(existing) + ConvertBytesToUint64(_new))
	}
	m := r.db.GetMergeOperator([]byte(key), add, 200*time.Millisecond)
	defer m.Stop()
	err = m.Add(convertUint64ToBytes(1))
	if err != nil {
		return 0, err
	}
	var val []byte
	val, err = m.Get()
	if err != nil {
		return 0, err
	}
	result = ConvertBytesToUint64(val)
	return result, nil
}

// GetSet получает значение по ключу и устанавливает новое значение
func (r *Repo) GetSet(key string, fn func(val string) (string, error)) (string, error) {
	var (
		val string
		err error
	)
	err = r.db.Update(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get([]byte(key))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		var valBytes []byte
		if err != badger.ErrKeyNotFound {
			valBytes, err = item.ValueCopy(nil)
			if err != nil {
				return err
			}
		}
		val, err = fn(string(valBytes))
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), []byte(val))
	})
	if err != nil {
		return "", err
	}
	return val, nil
}

// Get получает значение по ключу
func (r *Repo) Get(key string) (string, error) {
	var (
		val string
		err error
	)
	err = r.db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get([]byte(key))
		if err != nil {
			return err
		}
		var valBytes []byte
		valBytes, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		val = string(valBytes)
		return nil
	})
	return val, err
}

// Set устанавливает значение по ключу
func (r *Repo) Set(key, val string) error {
	err := r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(val))
	})
	return err
}

// Delete удаляет значение по ключу
func (r *Repo) Delete(key string) error {
	err := r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	return err
}

// convertUint64ToBytes преобразует uint64 в байтовый массив
func convertUint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

// ConvertBytesToUint64 преобразует байтовый массив в uint64
func ConvertBytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

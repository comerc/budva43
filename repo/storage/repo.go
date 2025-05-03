package storage

import (
	"context"
	"encoding/binary"
	"log/slog"
	"time"

	badger "github.com/dgraph-io/badger/v4"

	"github.com/comerc/budva43/config"
)

// Repo определяет интерфейс для работы с хранилищем BadgerDB
type Repo struct {
	log *slog.Logger
	//
	db *badger.DB
}

// New создает новый экземпляр репозитория для BadgerDB
func New() *Repo {
	return &Repo{
		log: slog.With("module", "repo.storage"),
		//
		db: nil,
	}
}

// Start устанавливает соединение с базой данных
func (r *Repo) Start(ctx context.Context, shutdown func()) error {
	opts := badger.DefaultOptions(config.Storage.DatabaseDirectory)
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
		return r.db.Close()
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
		again:
			err := r.db.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	}
}

// Increment увеличивает значение по ключу на 1
func (r *Repo) Increment(key string) (string, error) {
	var (
		err error
		val string
	)
	defer r.logOperation(&err, "Increment", key, &val)
	// Merge function to add two uint64 numbers
	add := func(existing, _new []byte) []byte {
		return convertUint64ToBytes(ConvertBytesToUint64(existing) + ConvertBytesToUint64(_new))
	}
	m := r.db.GetMergeOperator([]byte(key), add, 200*time.Millisecond)
	defer m.Stop()
	err = m.Add(convertUint64ToBytes(1))
	if err != nil {
		return "", err
	}
	var valBytes []byte
	valBytes, err = m.Get()
	if err != nil {
		return "", err
	}
	val = string(valBytes)
	return val, nil
}

// Get получает значение по ключу
func (r *Repo) Get(key string) (string, error) {
	var (
		err error
		val string
	)
	defer r.logOperation(&err, "Get", key, &val)
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
	var err error
	defer r.logOperation(&err, "Set", key, &val)
	err = r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(val))
	})
	return err
}

// Delete удаляет значение по ключу
func (r *Repo) Delete(key string) error {
	var err error
	defer r.logOperation(&err, "Delete", key, nil)
	err = r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	return err
}

// logOperation логирует операцию
func (r *Repo) logOperation(errPointer *error, name string, key string, val *string) {
	err := *errPointer
	if err == nil {
		if val != nil {
			r.log.Info(name, "key", key, "val", *val)
		} else {
			r.log.Info(name, "key", key)
		}
	} else {
		r.log.Error(name, "key", key, "err", err)
	}
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

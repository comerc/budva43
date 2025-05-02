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
func (r *Repo) Increment(key []byte) ([]byte, error) {
	var (
		err error
		val []byte
	)
	defer func() {
		if err != nil {
			r.log.Error("Increment", "key", key, "err", err)
		} else {
			r.log.Info("Increment", "key", key, "val", val)
		}
	}()
	// Merge function to add two uint64 numbers
	add := func(existing, _new []byte) []byte {
		return convertUint64ToBytes(ConvertBytesToUint64(existing) + ConvertBytesToUint64(_new))
	}
	m := r.db.GetMergeOperator(key, add, 200*time.Millisecond)
	defer m.Stop()
	err = m.Add(convertUint64ToBytes(1))
	if err != nil {
		return nil, err
	}
	val, err = m.Get()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Get получает значение по ключу
func (r *Repo) Get(key string) (string, error) {
	var (
		err error
		val []byte
	)
	defer func() {
		if err != nil {
			r.log.Error("Get", "key", key, "err", err)
		} else {
			r.log.Info("Get", "key", key, "val", val)
		}
	}()
	err = r.db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	return string(val), err
}

// Set устанавливает значение по ключу
func (r *Repo) Set(key, val string) error {
	var err error
	defer func() {
		if err != nil {
			r.log.Error("Set", "key", key, "err", err)
		} else {
			r.log.Info("Set", "key", key, "val", val)
		}
	}()
	err = r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(val))
	})
	return err
}

// Delete удаляет значение по ключу
func (r *Repo) Delete(key string) error {
	var err error
	defer func() {
		if err != nil {
			r.log.Error("Delete", "key", key, "err", err)
		} else {
			r.log.Info("Delete", "key", key)
		}
	}()
	err = r.db.Update(func(txn *badger.Txn) error {
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

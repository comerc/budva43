package badger

import (
	"context"
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
		log: slog.With("module", "repo.badger"),
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
	return nil
}

// Stop закрывает соединение с базой данных
func (r *Repo) Stop() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Get получает значение по ключу
func (r *Repo) Get(key string) ([]byte, error) {
	var value []byte
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

// Set устанавливает значение по ключу
func (r *Repo) Set(key string, value []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// SetWithTTL устанавливает значение по ключу с временем жизни
func (r *Repo) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	return r.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

// Delete удаляет значение по ключу
func (r *Repo) Delete(key string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Iterate выполняет итерацию по всем ключам с заданным префиксом
func (r *Repo) Iterate(prefix string, fn func(key string, value []byte) error) error {
	return r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				return fn(string(key), val)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

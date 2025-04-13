package badger

import (
	"context"
	"time"

	"github.com/comerc/budva43/config"
	badger "github.com/dgraph-io/badger/v4"
)

// Repo определяет интерфейс для работы с хранилищем BadgerDB
type Repo struct {
	db *badger.DB
}

// New создает новый экземпляр репозитория для BadgerDB
func New() *Repo {
	return &Repo{}
}

// Start устанавливает соединение с базой данных
func (r *Repo) Start(ctx context.Context, cancel context.CancelFunc) error {
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
func (r *Repo) Get(key []byte) ([]byte, error) {
	var value []byte
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

// Set устанавливает значение по ключу
func (r *Repo) Set(key, value []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// SetWithTTL устанавливает значение по ключу с временем жизни
func (r *Repo) SetWithTTL(key, value []byte, ttl time.Duration) error {
	return r.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

// Delete удаляет значение по ключу
func (r *Repo) Delete(key []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Iterate выполняет итерацию по всем ключам с заданным префиксом
func (r *Repo) Iterate(prefix []byte, fn func(key, value []byte) error) error {
	return r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				return fn(key, val)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

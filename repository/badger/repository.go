package badger

import (
	"context"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// Repository определяет интерфейс для работы с хранилищем BadgerDB
type Repository struct {
	db     *badger.DB
	dbPath string
}

// New создает новый экземпляр репозитория для BadgerDB
func New(dbPath string) *Repository {
	return &Repository{
		dbPath: dbPath,
	}
}

// Connect устанавливает соединение с базой данных
func (r *Repository) Connect(ctx context.Context) error {
	opts := badger.DefaultOptions(r.dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	r.db = db
	return nil
}

// Close закрывает соединение с базой данных
func (r *Repository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Get получает значение по ключу
func (r *Repository) Get(key []byte) ([]byte, error) {
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
func (r *Repository) Set(key, value []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// SetWithTTL устанавливает значение по ключу с временем жизни
func (r *Repository) SetWithTTL(key, value []byte, ttl time.Duration) error {
	return r.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

// Delete удаляет значение по ключу
func (r *Repository) Delete(key []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Iterate выполняет итерацию по всем ключам с заданным префиксом
func (r *Repository) Iterate(prefix []byte, fn func(key, value []byte) error) error {
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

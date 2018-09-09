package storage

import (
	"time"

	"github.com/datskos/ratelimit/pkg/config"
	"github.com/dgraph-io/badger"
)

type Storage interface {
	Tx() StorageTx
	Get(key string) (*Value, error)
	Close()
}

// Use Storage.Tx() when you need a create a transaction
// that spans multiple get/set calls
type StorageTx interface {
	Get(key string) (*Value, error)
	Set(key string, value *Value) error
	Commit() error
}

type Value struct {
	Remaining      uint32
	LastRefilledAt time.Time
	LastReducedAt  time.Time
}

type storage struct {
	db *badger.DB
}

func NewStorage(config config.AppConfig) (Storage, error) {
	opts := badger.DefaultOptions
	opts.Dir = config.DatabaseDir
	opts.ValueDir = config.DatabaseDir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &storage{db}, nil
}

func (storage *storage) Tx() StorageTx {
	return &tx{storage.db.NewTransaction(true)}
}

func (storage *storage) Get(key string) (*Value, error) {
	tx := storage.Tx()
	defer tx.Commit()
	return tx.Get(key)
}

func (storage *storage) Close() {
	storage.db.Close()
}

type tx struct {
	txn *badger.Txn
}

func (tx *tx) Get(key string) (*Value, error) {
	item, err := tx.txn.Get([]byte(key))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	data, err := item.Value()
	if err != nil {
		return nil, err
	}

	return decode(data)
}

func (tx *tx) Set(key string, value *Value) error {
	data, err := encode(value)
	if err != nil {
		return err
	}

	err = tx.txn.Set([]byte(key), data)
	if err != nil {
		return err
	}

	return nil
}

func (tx *tx) Commit() error {
	return tx.txn.Commit(nil)
}

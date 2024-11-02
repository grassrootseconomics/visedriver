package common

import (
	"git.defalsify.org/vise.git/db"

	"git.grassecon.net/urdt/ussd/internal/storage"
)

func StoreToDb(store *UserDataStore) db.Db {
	return store.Db
}

func StoreToPrefixDb(store *UserDataStore, pfx []byte) storage.PrefixDb {
	return storage.NewSubPrefixDb(store.Db, pfx)	
}

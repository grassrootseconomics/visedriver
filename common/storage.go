package common

import (
	"git.defalsify.org/vise.git/db"
)

func StoreToDb(store *UserDataStore, prefix []byte) db.Db {
	innerStore := store.Db
	if pfx != nil {
		innerStore = NewSubPrefixDb(innerStore, pfx)		
	}
	return innerStore
}

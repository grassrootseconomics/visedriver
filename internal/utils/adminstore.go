package utils

import (
	"context"
	"encoding/json"
	"os"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/logging"
)

var (
	logg = logging.NewVanilla().WithDomain("adminstore")
)

type Admin struct {
	PhoneNumber string `json:"phonenumber"`
}

type Config struct {
	Admins []Admin `json:"admins"`
}

type AdminStore struct {
	ctx     context.Context
	fsStore db.Db
}

func NewAdminStore(ctx context.Context, fileName string) (*AdminStore, error) {
	fsStore, err := getFsStore(ctx, fileName)
	if err != nil {
		return nil, err
	}
	return &AdminStore{ctx: ctx, fsStore: fsStore}, nil
}

func getFsStore(ctx context.Context, connectStr string) (db.Db, error) {
	fsStore := fsdb.NewFsDb()
	err := fsStore.Connect(ctx, connectStr)
	fsStore.SetPrefix(db.DATATYPE_USERDATA)
	if err != nil {
		return nil, err
	}
	return fsStore, nil
}

// Seed initializes a list of phonenumbers with admin privileges
func (as *AdminStore) Seed() error {
	var config Config

	store := as.fsStore
	defer store.Close()

	data, err := os.ReadFile("admin_numbers.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	for _, admin := range config.Admins {
		err := store.Put(as.ctx, []byte(admin.PhoneNumber), []byte("1"))
		if err != nil {
			logg.Printf(logging.LVL_DEBUG, "Failed to insert admin number", admin.PhoneNumber)
			return err
		}
	}
	return nil
}

// Checks if the given sessionId is listed as an admin.
func (as *AdminStore) IsAdmin(sessionId string) (bool, error) {
	_, err := as.fsStore.Get(as.ctx, []byte(sessionId))
	if err != nil {
		if db.IsNotFound(err) {
			logg.Printf(logging.LVL_INFO, "Returning false because session id was not found")
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

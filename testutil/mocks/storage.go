package mocks

import (
	"context"

	"git.defalsify.org/vise.git/db"
	memdb "git.defalsify.org/vise.git/db/mem"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
)

type MemStorageService struct {
	Db db.Db
	pe *persist.Persister
	rs resource.Resource
}

func NewMemStorageService(ctx context.Context) *MemStorageService {
	svc := &MemStorageService{
		Db: memdb.NewMemDb(),
	}
	err := svc.Db.Connect(ctx, "")
	if err != nil {
		panic(err)
	}
	svc.pe = persist.NewPersister(svc.Db)
	svc.rs = resource.NewMenuResource()
	return svc
}

func (mss *MemStorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	return mss.pe, nil
}

func (mss *MemStorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	return mss.Db, nil
}

func (mss *MemStorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	return mss.rs, nil
}

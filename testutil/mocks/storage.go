package mocks

import (
	"context"

	"github.com/grassrootseconomics/go-vise/db"
	memdb "github.com/grassrootseconomics/go-vise/db/mem"
	"github.com/grassrootseconomics/go-vise/persist"
	"github.com/grassrootseconomics/go-vise/resource"
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

package ussd

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"testing"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/testdata"
	testdataloader "github.com/peteole/testdata-loader"
)

var (
	dataGenerated bool   = false
	dataDir       string = testdata.DataDir
	BaseDir              = testdataloader.GetBasePath()
)

type testWrapper struct {
	resource.Resource
	db db.Db
}

func generateTestData(t *testing.T) {
	if dataGenerated {
		return
	}
	var err error
	dataDir, err = testdata.Generate()
	if err != nil {
		t.Fatal(err)
	}
}

func newTestWrapper(path string) testWrapper {
	ctx := context.Background()
	store := fsdb.NewFsDb()
	store.Connect(ctx, path)
	rs := resource.NewDbResource(store)
	rs.With(db.DATATYPE_STATICLOAD)
	wr := testWrapper{
		rs,
		store,
	}
	rs.AddLocalFunc("quit", quit)

	return wr
}

func quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "Thank you for using Sarafu network",
	}, nil
}

func TestTerms(t *testing.T) {
	generateTestData(t)
	ctx := context.Background()
	rs := newTestWrapper(dataDir)
	cfg := engine.Config{
		Root:      "terms",
		FlagCount: uint32(9),
	}
	store := storage.NewThreadGdbmDb()
	storeFile := path.Join(baseDir, "state.gdbm")
	err := store.Connect(ctx, storeFile)
	if err != nil {
		t.Fail()
	}

	pr := persist.NewPersister(store)
	en := engine.NewEngine(cfg, &rs)
	en.WithPersister(pr)
	if pr.GetState() == nil || pr.GetMemory() == nil {
		t.Fail()
	}
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	w := bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	b := w.Bytes()

	expect_str := `Do you agree to terms and conditions?
1:yes
2:no`

	if !bytes.Equal(b, []byte(expect_str)) {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect_str, b)
	}

	tests := []struct {
		name           string
		expectedSymbol string
		input          []byte
	}{
		{
			name:           "Test accept terms option(yes)",
			expectedSymbol: "create_pin",
			input:          []byte("1"),
		},
		// {
		// 	name:           "Test reject terms option(no)",
		// 	input:          []byte("2"),
		// 	expectedSymbol: "quit",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = en.Exec(ctx, tt.input)
			if err != nil {
				t.Fatal(err)
			}
			w := bytes.NewBuffer(nil)
			_, err = en.Flush(ctx, w)
			if err != nil {
				t.Fatal(err)
			}

			b = w.Bytes()
			fmt.Println("result", string(b))
			symbol, _ := pr.State.Where()
			
			if symbol != tt.expectedSymbol {
				t.Fatalf("expected symbol to be 'create_pin', got %s", symbol)
			}
		})
	}
}

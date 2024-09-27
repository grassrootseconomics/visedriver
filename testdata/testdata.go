package testdata

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/vm"
	testdataloader "github.com/peteole/testdata-loader"
)

func outNew(sym string, b []byte, tpl string, data map[string]string) error {
	logg.Debugf("testdata out", "sym", sym)
	store.SetPrefix(db.DATATYPE_TEMPLATE)
	err := store.Put(ctx, []byte(sym), []byte(tpl))
	if err != nil {
		return err
	}
	store.SetPrefix(db.DATATYPE_BIN)
	err = store.Put(ctx, []byte(sym), b)
	if err != nil {
		return err
	}
	store.SetPrefix(db.DATATYPE_STATICLOAD)
	for k, v := range data {
		logg.Debugf("testdata out staticload", "sym", sym, "k", k, "v", v)
		err = store.Put(ctx, []byte(k), []byte(v))
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	ctx     = context.Background()
	store   = fsdb.NewFsDb()
	out     = outNew
	logg    = logging.NewVanilla().WithDomain("testdata")
	BaseDir = testdataloader.GetBasePath()
	DataDir = ""
	dirLock = false
)

type genFunc func() error

func terms() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"yes", "1"}, nil, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"no", "2"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"create_pin", "1"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"quit", "2"}, nil, nil)
	tpl := "Do you agree to terms and conditions?"
	return out("terms", b, tpl, nil)
}

func createPin() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"exit", "0"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"0", "1"}, nil, nil)
	tpl := "create pin"
	return out("create_pin", b, tpl, nil)
}

func quit() error {
	// b := []byte{}
	// b = vm.NewLine(b, vm.LOAD, []string{"quit"}, []byte{0x00}, nil)
	// //b = vm.NewLine(b, vm.RELOAD, []string{"quit"}, []byte{0x00}, nil)
	// b = vm.NewLine(b, vm.HALT, nil, nil, nil)

	// return out("quit", b, "quit", nil)
	b := vm.NewLine(nil, vm.LOAD, []string{"quit"}, []byte{0x00}, nil)
	b = vm.NewLine(b, vm.RELOAD, []string{"quit"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)

	fp := path.Join(DataDir, "nothing.bin")
	err := os.WriteFile(fp, b, 0600)
	return err
}

func generate() error {
	err := os.MkdirAll(DataDir, 0755)
	if err != nil {
		return err
	}
	store = fsdb.NewFsDb()
	store.Connect(ctx, DataDir)
	store.SetLock(db.DATATYPE_TEMPLATE, false)
	store.SetLock(db.DATATYPE_BIN, false)
	store.SetLock(db.DATATYPE_MENU, false)
	store.SetLock(db.DATATYPE_STATICLOAD, false)

	fns := []genFunc{terms, createPin, quit}
	for _, fn := range fns {
		err = fn()
		if err != nil {
			return err
		}
	}
	return nil
}

func Generate() (string, error) {
	dir, err := ioutil.TempDir("", "vise_testdata_")
	if err != nil {
		return "", err
	}
	DataDir = dir
	dirLock = true
	err = generate()
	return dir, err
}

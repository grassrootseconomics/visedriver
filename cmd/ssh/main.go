package main

import (
	"context"
	"flag"
	"fmt"
	"path"
	"os"
	"sync"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"

	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/internal/ssh"
)

var (
	wg sync.WaitGroup
	keyStore db.Db
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

func main() {
	var dbDir string
	var resourceDir string
	var size uint
	var engineDebug bool
	var stateDebug bool
	var host string
	var port uint
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.BoolVar(&engineDebug, "engine-debug", false, "use engine debug output")
	flag.BoolVar(&stateDebug, "state-debug", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.StringVar(&host, "h", "127.0.0.1", "http host")
	flag.UintVar(&port, "p", 7122, "http port")
	flag.Parse()

	sshKeyFile := flag.Arg(0)
	_, err := os.Stat(sshKeyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open ssh server private key file: %v\n", err)
		os.Exit(1)
	}

	logg.Infof("start command", "dbdir", dbDir, "resourcedir", resourceDir, "outputsize", size, "keyfile", sshKeyFile, "host", host, "port", port)

	ctx := context.Background()
	pfp := path.Join(scriptDir, "pp.csv")

	cfg := engine.Config{
		Root:       "root",
		OutputSize: uint32(size),
		FlagCount:  uint32(16),
	}
	if stateDebug {
		cfg.StateDebug = true
	}
	if engineDebug {
		cfg.EngineDebug = true
	}

	keyStoreFile := path.Join(dbDir, "ssh_authorized_keys.gdbm")
	authKeyStore := storage.NewThreadGdbmDb()
	err = authKeyStore.Connect(ctx, keyStoreFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keystore file open error: %v", err)
		os.Exit(1)
	}
	defer func() {
		err := authKeyStore.Close()
		if err != nil {
			logg.ErrorCtxf(ctx, "keystore close error", "err", err)
		}
	}()

	runner := &ssh.SshRunner{
		Cfg: cfg,
		Debug: engineDebug,
		FlagFile: pfp,
		DbDir: dbDir,
		ResourceDir: resourceDir,
		SrvKeyFile: sshKeyFile,
		Host: host,
		Port: port,
	}
	runner.Run(ctx, authKeyStore)
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"path"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"

	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
)

var (
	wg sync.WaitGroup
	auth map[string]string
	keyStore db.Db
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

type auther struct {
	SessionId string
	Ctx context.Context
}

func(a *auther) Check(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	keyStore.SetPrefix(storage.DATATYPE_CUSTOM)
	k := append([]byte{0x01}, pubKey.Marshal()...)
	v, err := keyStore.Get(a.Ctx, k)
	if err != nil {
		return nil, err
	}
	a.SessionId = string(v)
	fmt.Fprintf(os.Stderr, "connect: %s\n", a.SessionId)
	return nil, nil
}

func populateAuth() {
	auth = make(map[string]string)
	pubKey, _, _, rest, err := ssh.ParseAuthorizedKey([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCu5rYCxMBsVAL1TEkMQgmElAYEZj5zYDdyHjUxZ6qzHBOZD9GAzdxx9GyQDx2vdYm3329tLH/69ky1YA3nUz8SnJGBD6hC5XrqwN6zo9R9oOHAKTwiPGhey2NTVmheP+9XNHukBnOlkkWOQlpDDvMbWOztaZOWDaA8OIeP0t6qzFqLyelyg65lxzM3BKd7bCmmfzl/64BcP1MotAmB9DUxmY0Wb4Q2hYZfNYBx50Z4xthTgKV+Xoo8CbTduKotIz6hluQGvWdtxlCJQEiZ2f4RYY87JSA6/BAH2fhxuLHMXRpzocJNqARqCWpdcTGSg7bzxbKvTFH9OU4wZtr9ie40OR4zsc1lOBZL0rnp8GLkG8ZmeBQrgEDlmR9TTlz4okgtL+c5TCS37rjZYVjmtGwihws0EL9+wyv2dSQibirklC4wK5eWHKXl5vab19qzw/qRLdoRBK40DxbRKggxA7gqSsKrmrf+z7CuLIz/kxF+169FBLbh1MfBOGdx1awm6aU= lash@furioso"))
	if err != nil {
		panic(err)
	}
	auth[string(pubKey.Marshal())] = "+25113243546"
	_ = rest
}

func serve(ch ssh.NewChannel) {
	if ch == nil {
		return
	}
	if ch.ChannelType() != "session" {
		ch.Reject(ssh.UnknownChannelType, "that is not the channel you are looking for")
		return
	}
	channel, requests, err := ch.Accept()
	if err != nil {
		panic(err)
	}
	defer channel.Close()
	wg.Add(1)
	go func(reqIn <-chan *ssh.Request) {
		defer wg.Done()
		for req := range reqIn {
			req.Reply(req.Type == "shell", nil)	
		}
		_ = requests
	}(requests)

	n, err := channel.Write([]byte("foobarbaz\n"))
	if err != nil {
		panic(err)
	}
	log.Printf("wrote %d", n)
}

func sshRun(ctx context.Context, hl *ussd.Handlers) {
	running := true

	defer wg.Wait()

	auth := auther{Ctx: ctx}
	cfg := ssh.ServerConfig{
		PublicKeyCallback: auth.Check,
	}

	privateBytes, err := os.ReadFile("/home/lash/.ssh/id_rsa_tmp")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	cfg.AddHostKey(private)

	lst, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		panic(err)
	}

	for running {
		conn, err := lst.Accept()
		if err != nil {
			panic(err)
		}

		go func(conn net.Conn) {
			defer conn.Close()
			for true {
				srvConn, nC, rC, err := ssh.NewServerConn(conn, &cfg)
				if err != nil {
					log.Printf("rejected client: %v", err)
				}
				log.Printf("haveconn %v", srvConn)

				wg.Add(1)
				go func() {
					ssh.DiscardRequests(rC)
					wg.Done()
				}()

				for ch := range nC {
					serve(ch)
				}
			}
		}(conn)
	}
}

func sshLoadKeys(ctx context.Context, dbDir string) error {
	keyStoreFile := path.Join(dbDir, "ssh_authorized_keys.gdbm")
	keyStore = gdbmdb.NewGdbmDb()
	keyStore.Connect(ctx, keyStoreFile)
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCu5rYCxMBsVAL1TEkMQgmElAYEZj5zYDdyHjUxZ6qzHBOZD9GAzdxx9GyQDx2vdYm3329tLH/69ky1YA3nUz8SnJGBD6hC5XrqwN6zo9R9oOHAKTwiPGhey2NTVmheP+9XNHukBnOlkkWOQlpDDvMbWOztaZOWDaA8OIeP0t6qzFqLyelyg65lxzM3BKd7bCmmfzl/64BcP1MotAmB9DUxmY0Wb4Q2hYZfNYBx50Z4xthTgKV+Xoo8CbTduKotIz6hluQGvWdtxlCJQEiZ2f4RYY87JSA6/BAH2fhxuLHMXRpzocJNqARqCWpdcTGSg7bzxbKvTFH9OU4wZtr9ie40OR4zsc1lOBZL0rnp8GLkG8ZmeBQrgEDlmR9TTlz4okgtL+c5TCS37rjZYVjmtGwihws0EL9+wyv2dSQibirklC4wK5eWHKXl5vab19qzw/qRLdoRBK40DxbRKggxA7gqSsKrmrf+z7CuLIz/kxF+169FBLbh1MfBOGdx1awm6aU= lash@furioso"))
	if err != nil {
		return err
	}
	//auth[string(pubKey.Marshal())] = "+25113243546"
	k := append([]byte{0x01}, pubKey.Marshal()...)
	keyStore.SetPrefix(storage.DATATYPE_CUSTOM)
	return keyStore.Put(ctx, k, []byte("+25113243546"))
}

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
	flag.UintVar(&port, "p", 7123, "http port")
	flag.Parse()

	logg.Infof("start command", "dbdir", dbDir, "resourcedir", resourceDir, "outputsize", size)

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

	menuStorageService := storage.MenuStorageService{}
	rs, err := menuStorageService.GetResource(scriptDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = menuStorageService.EnsureDbDir(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	userdataStore := menuStorageService.GetUserdataDb(dbDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer userdataStore.Close()

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		os.Exit(1)
	}

	lhs, err := handlers.NewLocalHandlerService(pfp, true, dbResource, cfg, rs)
	lhs.WithDataStore(&userdataStore)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	hl, err := lhs.GetHandler()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	stateStore, err := menuStorageService.GetStateStore(dbDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer stateStore.Close()

	sshLoadKeys(ctx, dbDir)
	sshRun(ctx, hl)
}

package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net"
	"path"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"

	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

var (
	wg sync.WaitGroup
	keyStore db.Db
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

type auther struct {
	Ctx context.Context
	auth map[string]string
}

func NewAuther(ctx context.Context) *auther {
	return &auther{
		Ctx: ctx,
		auth: make(map[string]string),
	}
}

func(a *auther) Check(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	keyStore.SetLanguage(nil)
	keyStore.SetPrefix(storage.DATATYPE_CUSTOM)
	k := append([]byte{0x01}, pubKey.Marshal()...)
	v, err := keyStore.Get(a.Ctx, k)
	if err != nil {
		return nil, err
	}
	ka := hex.EncodeToString(conn.SessionID())
	va := string(v)
	a.auth[ka] = va 
	fmt.Fprintf(os.Stderr, "connect: %s -> %s\n", ka, v)
	return nil, nil
}

func(a *auther) FromConn(c *ssh.ServerConn) (string, error) {
	if c == nil {
		return "", errors.New("nil server conn")
	}
	if c.Conn == nil {
		return "", errors.New("nil underlying conn")
	}
	return a.Get(c.Conn.SessionID())
}


func(a *auther) Get(k []byte) (string, error) {
	ka := hex.EncodeToString(k)
	v, ok := a.auth[ka]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

//func serve(ctx context.Context, sessionId string, ch ssh.NewChannel, mss *storage.MenuStorageService, lhs *handlers.LocalHandlerService) error {
func serve(ctx context.Context, sessionId string, ch ssh.NewChannel, en engine.Engine) error {
	if ch == nil {
		return errors.New("nil channel")
	}
	if ch.ChannelType() != "session" {
		ch.Reject(ssh.UnknownChannelType, "that is not the channel you are looking for")
		return errors.New("not a session")
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

	cont, err := en.Exec(ctx, []byte{})
	if err != nil {
		return fmt.Errorf("initial engine exec err: %v", err)
	}

	var input [state.INPUT_LIMIT]byte
	for cont {
		c, err := en.Flush(ctx, channel)
		if err != nil {
			return fmt.Errorf("flush err: %v", err)
		}
		_, err = channel.Write([]byte{0x0a})
		if err != nil {
			return fmt.Errorf("newline err: %v", err)
		}
		c, err = channel.Read(input[:])
		if err != nil {
			return fmt.Errorf("read input fail: %v", err)
		}
		logg.TraceCtxf(ctx, "input read", "c", c, "input", input[:c-1])
		cont, err = en.Exec(ctx, input[:c-1])
		if err != nil {
			return fmt.Errorf("engine exec err: %v", err)
		}
		logg.TraceCtxf(ctx, "exec cont", "cont", cont, "en", en)
		_ = c
	}
	c, err := en.Flush(ctx, channel)
	if err != nil {
		return fmt.Errorf("last flush err: %v", err)
	}
	_ = c
	return nil
}

type sshRunner struct {
	Ctx context.Context
	Cfg engine.Config
	FlagFile string
	DbDir string
	ResourceDir string
	Debug bool
	KeyFile string
	Host string
	Port uint
}

func(s *sshRunner) GetEngine(sessionId string) (engine.Engine, func(), error) {
	ctx := s.Ctx
	menuStorageService := storage.NewMenuStorageService(s.DbDir, s.ResourceDir)

	err := menuStorageService.EnsureDbDir()
	if err != nil {
		return nil, nil, err
	}

	rs, err := menuStorageService.GetResource(ctx)
	if err != nil {
		return nil, nil, err
	}

	pe, err := menuStorageService.GetPersister(ctx)
	if err != nil {
		return nil, nil, err
	}

	userdatastore, err := menuStorageService.GetUserdataDb(ctx)
	if err != nil {
		return nil, nil, err
	}

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		return nil, nil, err
	}

	lhs, err := handlers.NewLocalHandlerService(s.FlagFile, true, dbResource, s.Cfg, rs)
	lhs.SetDataStore(&userdatastore)
	lhs.SetPersister(pe)
	lhs.Cfg.SessionId = sessionId

	if err != nil {
		return nil, nil, err
	}

	hl, err := lhs.GetHandler()
	if err != nil {
		return nil, nil, err
	}

	en := lhs.GetEngine()
	en = en.WithFirst(hl.Init)
	if s.Debug {
		en = en.WithDebug(nil)
	}
	// TODO: this is getting very hacky!
	closer := func() {
		err := menuStorageService.Close()
		if err != nil {
			logg.ErrorCtxf(ctx, "menu storage service cleanup fail", "err", err)
		}
	}
	return en, closer, nil
}

// adapted example from crypto/ssh package, NewServerConn doc
func(s *sshRunner) Run(ctx context.Context) {
	running := true

	// TODO: waitgroup should probably not be global
	defer wg.Wait()

	auth := NewAuther(ctx)
	cfg := ssh.ServerConfig{
		PublicKeyCallback: auth.Check,
	}

	privateBytes, err := os.ReadFile(s.KeyFile)
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to load private key", "err", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to parse private key", "err", err)
	}
	cfg.AddHostKey(private)

	lst, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
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
					logg.InfoCtxf(ctx, "rejected client", "err", err)
					return
				}
				logg.DebugCtxf(ctx, "ssh client connected", "conn", srvConn)

				wg.Add(1)
				go func() {
					ssh.DiscardRequests(rC)
					wg.Done()
				}()
				
				sessionId, err := auth.FromConn(srvConn)
				if err != nil {
					logg.ErrorCtxf(ctx, "Cannot find authentication")
					return
				}
				en, closer, err := s.GetEngine(sessionId)
				if err != nil {
					logg.ErrorCtxf(ctx, "engine won't start", "err", err)
					return
				}
				defer func() {
					err := en.Finish()
					if err != nil {
						logg.ErrorCtxf(ctx, "engine won't stop", "err", err)
					}
					closer()
				}()
				for ch := range nC {
					err = serve(ctx, sessionId, ch, en)
					logg.ErrorCtxf(ctx, "ssh server finish", "err", err)
				}
			}
		}(conn)
	}
}

// TODO: This is test code, move to external tool for adding and removing keys
func sshLoadKeys(ctx context.Context, dbDir string) error {
	keyStoreFile := path.Join(dbDir, "ssh_authorized_keys.gdbm")
	keyStore = storage.NewThreadGdbmDb()
	err := keyStore.Connect(ctx, keyStoreFile)
	if err != nil {
		return err
	}
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCu5rYCxMBsVAL1TEkMQgmElAYEZj5zYDdyHjUxZ6qzHBOZD9GAzdxx9GyQDx2vdYm3329tLH/69ky1YA3nUz8SnJGBD6hC5XrqwN6zo9R9oOHAKTwiPGhey2NTVmheP+9XNHukBnOlkkWOQlpDDvMbWOztaZOWDaA8OIeP0t6qzFqLyelyg65lxzM3BKd7bCmmfzl/64BcP1MotAmB9DUxmY0Wb4Q2hYZfNYBx50Z4xthTgKV+Xoo8CbTduKotIz6hluQGvWdtxlCJQEiZ2f4RYY87JSA6/BAH2fhxuLHMXRpzocJNqARqCWpdcTGSg7bzxbKvTFH9OU4wZtr9ie40OR4zsc1lOBZL0rnp8GLkG8ZmeBQrgEDlmR9TTlz4okgtL+c5TCS37rjZYVjmtGwihws0EL9+wyv2dSQibirklC4wK5eWHKXl5vab19qzw/qRLdoRBK40DxbRKggxA7gqSsKrmrf+z7CuLIz/kxF+169FBLbh1MfBOGdx1awm6aU= lash@furioso"))
	if err != nil {
		return err
	}
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
	
	err = sshLoadKeys(ctx, dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	runner := &sshRunner{
		Cfg: cfg,
		Debug: engineDebug,
		FlagFile: pfp,
		DbDir: dbDir,
		ResourceDir: resourceDir,
		KeyFile: sshKeyFile,
		Host: host,
		Port: port,
	}
	runner.Run(ctx)
}

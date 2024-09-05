package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	fsdb "git.defalsify.org/vise.git/db/fs"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
)

var (
	logg = logging.NewVanilla()
	flags *FlagParser
)


func getFlags(fp string, debug bool) error {
	Flags = NewFlagParser().WithDebug()
	flags, err := Flags.Load(fp)
	if err != nil {
		return err
	}
}

func getHandler(appFlags *asm.FlagParser, rs *resource.DbResource, pe *persist.Persister, userdataStore db.Db) (*ussd.Handlers, error) {
	ussdHandlers, err := ussd.NewHandlers(appFlags pr.GetState(), dataStore)
	if err != nil {
		return nil, err
	}
	rs.AddLocalFunc("select_language", ussdHandlers.SetLanguage)
	rs.AddLocalFunc("create_account", ussdHandlers.CreateAccount)
	rs.AddLocalFunc("save_pin", ussdHandlers.SavePin)
	rs.AddLocalFunc("verify_pin", ussdHandlers.VerifyPin)
	rs.AddLocalFunc("check_identifier", ussdHandlers.CheckIdentifier)
	rs.AddLocalFunc("check_account_status", ussdHandlers.CheckAccountStatus)
	rs.AddLocalFunc("authorize_account", ussdHandlers.Authorize)
	rs.AddLocalFunc("quit", ussdHandlers.Quit)
	rs.AddLocalFunc("check_balance", ussdHandlers.CheckBalance)
	rs.AddLocalFunc("validate_recipient", ussdHandlers.ValidateRecipient)
	rs.AddLocalFunc("transaction_reset", ussdHandlers.TransactionReset)
	rs.AddLocalFunc("max_amount", ussdHandlers.MaxAmount)
	rs.AddLocalFunc("validate_amount", ussdHandlers.ValidateAmount)
	rs.AddLocalFunc("reset_transaction_amount", ussdHandlers.ResetTransactionAmount)
	rs.AddLocalFunc("get_recipient", ussdHandlers.GetRecipient)
	rs.AddLocalFunc("get_sender", ussdHandlers.GetSender)
	rs.AddLocalFunc("get_amount", ussdHandlers.GetAmount)
	rs.AddLocalFunc("reset_incorrect", ussdHandlers.ResetIncorrectPin)
	rs.AddLocalFunc("save_firstname", ussdHandlers.SaveFirstname)
	rs.AddLocalFunc("save_familyname", ussdHandlers.SaveFamilyname)
	rs.AddLocalFunc("save_gender", ussdHandlers.SaveGender)
	rs.AddLocalFunc("save_location", ussdHandlers.SaveLocation)
	rs.AddLocalFunc("save_yob", ussdHandlers.SaveYob)
	rs.AddLocalFunc("save_offerings", ussdHandlers.SaveOfferings)
	rs.AddLocalFunc("quit_with_balance", ussdHandlers.QuitWithBalance)
	rs.AddLocalFunc("reset_account_authorized", ussdHandlers.ResetAccountAuthorized)
	rs.AddLocalFunc("reset_allow_update", ussdHandlers.ResetAllowUpdate)
	rs.AddLocalFunc("get_profile_info", ussdHandlers.GetProfileInfo)
	rs.AddLocalFunc("verify_yob", ussdHandlers.VerifyYob)
	rs.AddLocalFunc("reset_incorrect_date_format", ussdHandlers.ResetIncorrectYob)
	rs.AddLocalFunc("set_reset_single_edit", ussdHandlers.SetResetSingleEdit)
	rs.AddLocalFunc("initiate_transaction", ussdHandlers.InitiateTransaction)

	return ussdHandlers, nil
}

func getDataPersister(dbDir string) (*persist.Persister, error) {
	err = os.MkdirAll(dp, 0700)
	if err != nil {
		return nil, fmt.Errorf("state dir create exited with error: %v\n", err)
	}

	dataStore := gdbmdb.NewGdbmDb()
	dataStoreFile := path.Join(dbDir, "states.gdbm")
	dataStore.Connect(ctx, dataStoreFile)
	pr := persist.NewPersister(dataStore)

	return pr
}

func getResource(resourceDir string) (resource.Resource, error) {
	store := fsdb.NewFsDb()
	err = store.Connect(ctx, resourceDir)
	if err != nil {
		return err
	}
	rfs := resource.NewDbResource(store)
}

func getEngine(cfg Config, rs resource.Resource, pr *persister.Persister) {
	cfg := engine.Config{
		Root:      sym,
		SessionId: sessionId,
		FlagCount: uint32(16),
	}
	en := engine.NewEngine(cfg, rfs)
	en = en.WithPersister(pr)
	return en
}	

func main() {
	var dbDir string
	var resourceDir string
	var root string
	var size uint
	var sessionId string
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	logg.Infof("starting session", "symbol", root, "dbdir", dbDir, "sessionid", sessionId, "outsize", size)

	fl, err := getFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	rs, err := getResource(resourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	pr, err := getDataPersister(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	store, err := getUserDb(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	hn, err := getHandlers(fl, rs, pr, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	en := getEngine(cfg, rs, pr)
	_, err = en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init exited with error: %v\n", err)
		os.Exit(1)
	}

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}

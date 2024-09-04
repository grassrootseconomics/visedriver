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

	"git.defalsify.org/vise.git/cache"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
)

var (
	scriptDir = path.Join("services", "registration")
	store     = fsdb.NewFsDb()
	pr        = persist.NewPersister(store)
)

type menuResource struct {
	*resource.DbResource
}

func newMenuResource(rs *resource.DbResource) resource.Resource {
	return &menuResource{
		rs,
	}
}

func main() {
	var dir string
	var root string
	var size uint
	var sessionId string
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()

	pfp := path.Join(scriptDir, "pp.csv")
	file, err := os.Open(pfp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open CSV file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	reader := csv.NewReader(file)

	// Iterate through the CSV records and register the flags
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading CSV file: %v\n", err)
			os.Exit(1)
		}

		// Ensure the record starts with "flag" and has at least 3 columns
		if len(record) < 3 || record[0] != "flag" {
			continue
		}

		flagName := record[1]
		flagValue, err := strconv.Atoi(record[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to convert flag value %s to integer: %v\n", record[2], err)
			continue
		}

		// Register the flag
		state.FlagDebugger.Register(uint32(flagValue), flagName)
	}

	ca := cache.NewCache()
	cfg := engine.Config{
		Root:      "root",
		SessionId: sessionId,
		FlagCount: uint32(16),
	}

	dp := path.Join(scriptDir, ".state")
	err = os.MkdirAll(dp, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "state dir create exited with error: %v\n", err)
		os.Exit(1)
	}

	store := fsdb.NewFsDb()
	err = store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}

	rfs := resource.NewDbResource(store)

	rs, ok := newMenuResource(rfs).(*menuResource)
	if !ok {
		os.Exit(1)
	}
	en := engine.NewEngine(cfg, rs)
	en = en.WithMemory(ca)
	en = en.WithPersister(pr)

	fp := path.Join(dp, sessionId)

	ussdHandlers, err := ussd.NewHandlers(fp, pr.State, sessionId)

	if err != nil {
		fmt.Fprintf(os.Stderr, "handler setup failed with error: %v\n", err)
	}

	rfs.AddLocalFunc("select_language", ussdHandlers.SetLanguage)
	rfs.AddLocalFunc("create_account", ussdHandlers.CreateAccount)
	rfs.AddLocalFunc("save_pin", ussdHandlers.SavePin)
	rfs.AddLocalFunc("verify_pin", ussdHandlers.VerifyPin)
	rfs.AddLocalFunc("check_identifier", ussdHandlers.CheckIdentifier)
	rfs.AddLocalFunc("check_account_status", ussdHandlers.CheckAccountStatus)
	rfs.AddLocalFunc("authorize_account", ussdHandlers.Authorize)
	rfs.AddLocalFunc("quit", ussdHandlers.Quit)
	rfs.AddLocalFunc("check_balance", ussdHandlers.CheckBalance)
	rfs.AddLocalFunc("validate_recipient", ussdHandlers.ValidateRecipient)
	rfs.AddLocalFunc("transaction_reset", ussdHandlers.TransactionReset)
	rfs.AddLocalFunc("max_amount", ussdHandlers.MaxAmount)
	rfs.AddLocalFunc("validate_amount", ussdHandlers.ValidateAmount)
	rfs.AddLocalFunc("reset_transaction_amount", ussdHandlers.ResetTransactionAmount)
	rfs.AddLocalFunc("get_recipient", ussdHandlers.GetRecipient)
	rfs.AddLocalFunc("get_sender", ussdHandlers.GetSender)
	rfs.AddLocalFunc("get_amount", ussdHandlers.GetAmount)
	rfs.AddLocalFunc("reset_incorrect", ussdHandlers.ResetIncorrectPin)
	rfs.AddLocalFunc("save_firstname", ussdHandlers.SaveFirstname)
	rfs.AddLocalFunc("save_familyname", ussdHandlers.SaveFamilyname)
	rfs.AddLocalFunc("save_gender", ussdHandlers.SaveGender)
	rfs.AddLocalFunc("save_location", ussdHandlers.SaveLocation)
	rfs.AddLocalFunc("save_yob", ussdHandlers.SaveYob)
	rfs.AddLocalFunc("save_offerings", ussdHandlers.SaveOfferings)
	rfs.AddLocalFunc("quit_with_balance", ussdHandlers.QuitWithBalance)
	rfs.AddLocalFunc("reset_account_authorized", ussdHandlers.ResetAccountAuthorized)
	rfs.AddLocalFunc("reset_allow_update", ussdHandlers.ResetAllowUpdate)
	rfs.AddLocalFunc("get_profile_info", ussdHandlers.GetProfileInfo)
	rfs.AddLocalFunc("verify_yob", ussdHandlers.VerifyYob)
	rfs.AddLocalFunc("reset_incorrect_date_format", ussdHandlers.ResetIncorrectYob)
	rfs.AddLocalFunc("set_reset_single_edit", ussdHandlers.SetResetSingleEdit)
	rfs.AddLocalFunc("initiate_transaction", ussdHandlers.InitiateTransaction)

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

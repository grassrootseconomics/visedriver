package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
	"git.grassecon.net/urdt/ussd/internal/models"
)

var (
	scriptDir = path.Join("services", "registration")
)

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
	st := state.NewState(16)
	st.UseDebug()
	state.FlagDebugger.Register(models.USERFLAG_LANGUAGE_SET, "LANGUAGE_CHANGE")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_CREATED, "ACCOUNT_CREATED")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_SUCCESS, "ACCOUNT_SUCCESS")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_PENDING, "ACCOUNT_PENDING")
	state.FlagDebugger.Register(models.USERFLAG_INCORRECTPIN, "INCORRECTPIN")
	state.FlagDebugger.Register(models.USERFLAG_INCORRECTDATEFORMAT, "INVALIDDATEFORMAT")
	state.FlagDebugger.Register(models.USERFLAG_INVALID_RECIPIENT, "INVALIDRECIPIENT")
	state.FlagDebugger.Register(models.USERFLAG_PINMISMATCH, "PINMISMATCH")
	state.FlagDebugger.Register(models.USERFLAG_PIN_SET, "PIN_SET")
	state.FlagDebugger.Register(models.USERFLAG_INVALID_RECIPIENT_WITH_INVITE, "INVALIDRECIPIENT_WITH_INVITE")
	state.FlagDebugger.Register(models.USERFLAG_INVALID_AMOUNT, "INVALIDAMOUNT")
	state.FlagDebugger.Register(models.USERFLAG_ALLOW_UPDATE, "UNLOCKFORUPDATE")
	state.FlagDebugger.Register(models.USERFLAG_VALIDPIN, "VALIDPIN")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_AUTHORIZED, "ACCOUNT_AUTHORIZED")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_CREATION_FAILED, "ACCOUNT_CREATION_FAILED")
	state.FlagDebugger.Register(models.USERFLAG_SINGLE_EDIT, "SINGLEEDIT")

	rfs := resource.NewFsResource(scriptDir)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root:      "root",
		SessionId: sessionId,
	}

	dp := path.Join(scriptDir, ".state")
	err := os.MkdirAll(dp, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "state dir create exited with error: %v\n", err)
		os.Exit(1)
	}
	pr := persist.NewFsPersister(dp)
	en, err := engine.NewPersistedEngine(ctx, cfg, pr, rfs)

	if err != nil {
		pr = pr.WithContent(&st, ca)
		err = pr.Save(cfg.SessionId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save state with error: %v\n", err)
		}
		en, err = engine.NewPersistedEngine(ctx, cfg, pr, rfs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "engine create exited with error: %v\n", err)
			os.Exit(1)
		}
	}

	fp := path.Join(dp, sessionId)

	ussdHandlers,err := ussd.NewHandlers(fp, &st)

	if(err != nil){
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

	cont, err := en.Init(ctx)
	en.SetDebugger(engine.NewSimpleDebug(nil))
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init exited with error: %v\n", err)
		os.Exit(1)
	}
	if !cont {
		_, err = en.WriteResult(ctx, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dead init write error: %v\n", err)
			os.Exit(1)
		}
		err = en.Finish()
		if err != nil {
			fmt.Fprintf(os.Stderr, "engine finish error: %v\n", err)
			os.Exit(1)
		}
		os.Stdout.Write([]byte{0x0a})
		os.Exit(0)
	}
	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}

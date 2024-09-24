package handlers

import (
	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
)

type HandlerService interface {
	GetHandler() (*ussd.Handlers, error)
}

func getParser(fp string, debug bool) (*asm.FlagParser, error) {
	flagParser := asm.NewFlagParser().WithDebug()
	_, err := flagParser.Load(fp)
	if err != nil {
		return nil, err
	}
	return flagParser, nil
}

type LocalHandlerService struct {
	Parser        *asm.FlagParser
	DbRs          *resource.DbResource
	Pe            *persist.Persister
	UserdataStore *db.Db
	Cfg           engine.Config
	Rs            resource.Resource
}

func NewLocalHandlerService(fp string, debug bool, dbResource *resource.DbResource, cfg engine.Config, rs resource.Resource) (*LocalHandlerService, error) {
	parser, err := getParser(fp, debug)
	if err != nil {
		return nil, err
	}
	return &LocalHandlerService{
		Parser: parser,
		DbRs:   dbResource,
		Cfg:    cfg,
		Rs:     rs,
	}, nil
}

func (ls *LocalHandlerService) SetPersister(Pe *persist.Persister) {
	ls.Pe = Pe
}

func (ls *LocalHandlerService) SetDataStore(db *db.Db) {
	ls.UserdataStore = db
}

func (ls *LocalHandlerService) GetHandler() (*ussd.Handlers, error) {
	ussdHandlers, err := ussd.NewHandlers(ls.Parser, *ls.UserdataStore)
	if err != nil {
		return nil, err
	}
	ussdHandlers = ussdHandlers.WithPersister(localHandlerService.Pe)
	localHandlerService.DbRs.AddLocalFunc("set_language", ussdHandlers.SetLanguage)
	localHandlerService.DbRs.AddLocalFunc("create_account", ussdHandlers.CreateAccount)
	localHandlerService.DbRs.AddLocalFunc("save_pin", ussdHandlers.SavePin)
	localHandlerService.DbRs.AddLocalFunc("verify_pin", ussdHandlers.VerifyPin)
	localHandlerService.DbRs.AddLocalFunc("check_identifier", ussdHandlers.CheckIdentifier)
	localHandlerService.DbRs.AddLocalFunc("check_account_status", ussdHandlers.CheckAccountStatus)
	localHandlerService.DbRs.AddLocalFunc("authorize_account", ussdHandlers.Authorize)
	localHandlerService.DbRs.AddLocalFunc("quit", ussdHandlers.Quit)
	localHandlerService.DbRs.AddLocalFunc("check_balance", ussdHandlers.CheckBalance)
	localHandlerService.DbRs.AddLocalFunc("validate_recipient", ussdHandlers.ValidateRecipient)
	localHandlerService.DbRs.AddLocalFunc("transaction_reset", ussdHandlers.TransactionReset)
	localHandlerService.DbRs.AddLocalFunc("max_amount", ussdHandlers.MaxAmount)
	localHandlerService.DbRs.AddLocalFunc("validate_amount", ussdHandlers.ValidateAmount)
	localHandlerService.DbRs.AddLocalFunc("reset_transaction_amount", ussdHandlers.ResetTransactionAmount)
	localHandlerService.DbRs.AddLocalFunc("get_recipient", ussdHandlers.GetRecipient)
	localHandlerService.DbRs.AddLocalFunc("get_sender", ussdHandlers.GetSender)
	localHandlerService.DbRs.AddLocalFunc("get_amount", ussdHandlers.GetAmount)
	localHandlerService.DbRs.AddLocalFunc("reset_incorrect", ussdHandlers.ResetIncorrectPin)
	localHandlerService.DbRs.AddLocalFunc("save_firstname", ussdHandlers.SaveFirstname)
	localHandlerService.DbRs.AddLocalFunc("save_familyname", ussdHandlers.SaveFamilyname)
	localHandlerService.DbRs.AddLocalFunc("save_gender", ussdHandlers.SaveGender)
	localHandlerService.DbRs.AddLocalFunc("save_location", ussdHandlers.SaveLocation)
	localHandlerService.DbRs.AddLocalFunc("save_yob", ussdHandlers.SaveYob)
	localHandlerService.DbRs.AddLocalFunc("save_offerings", ussdHandlers.SaveOfferings)
	localHandlerService.DbRs.AddLocalFunc("quit_with_balance", ussdHandlers.QuitWithBalance)
	localHandlerService.DbRs.AddLocalFunc("reset_account_authorized", ussdHandlers.ResetAccountAuthorized)
	localHandlerService.DbRs.AddLocalFunc("reset_allow_update", ussdHandlers.ResetAllowUpdate)
	localHandlerService.DbRs.AddLocalFunc("get_profile_info", ussdHandlers.GetProfileInfo)
	localHandlerService.DbRs.AddLocalFunc("verify_yob", ussdHandlers.VerifyYob)
	localHandlerService.DbRs.AddLocalFunc("reset_incorrect_date_format", ussdHandlers.ResetIncorrectYob)
	localHandlerService.DbRs.AddLocalFunc("initiate_transaction", ussdHandlers.InitiateTransaction)
	localHandlerService.DbRs.AddLocalFunc("save_temporary_pin", ussdHandlers.SaveTemporaryPin)
	localHandlerService.DbRs.AddLocalFunc("verify_new_pin", ussdHandlers.VerifyNewPin)
	localHandlerService.DbRs.AddLocalFunc("confirm_pin_change", ussdHandlers.ConfirmPinChange)
	localHandlerService.DbRs.AddLocalFunc("quit_with_help", ussdHandlers.QuitWithHelp)

	return ussdHandlers, nil
}

// TODO: enable setting of sessionId on engine init time
func (ls *LocalHandlerService) GetEngine() *engine.DefaultEngine {
	en := engine.NewEngine(ls.Cfg, ls.Rs)
	en = en.WithPersister(ls.Pe)
	return en
}

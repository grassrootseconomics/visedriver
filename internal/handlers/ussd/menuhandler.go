package ussd

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"git.defalsify.org/vise.git/asm"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/handlers/server"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"gopkg.in/leonelquinteros/gotext.v1"
)

var (
	scriptDir      = path.Join("services", "registration")
	translationDir = path.Join(scriptDir, "locale")
)

const (
	TrackingIdKey  = "TRACKINGID"
	PublicKeyKey   = "PUBLICKEY"
	CustodialIdKey = "CUSTODIALID"
	AccountPin     = "ACCOUNTPIN"
	AccountStatus  = "ACCOUNTSTATUS"
	FirstName      = "FIRSTNAME"
	FamilyName     = "FAMILYNAME"
	YearOfBirth    = "YOB"
	Location       = "LOCATION"
	Gender         = "GENDER"
	Offerings      = "OFFERINGS"
	Recipient      = "RECIPIENT"
	Amount         = "AMOUNT"
	AccountCreated = "ACCOUNTCREATED"
)

type FSData struct {
	Path string
	St   *state.State
}

// FlagManager handles centralized flag management
type FlagManager struct {
	parser *asm.FlagParser
}

// NewFlagManager creates a new FlagManager instance
func NewFlagManager(csvPath string) (*FlagManager, error) {
	parser := asm.NewFlagParser()
	_, err := parser.Load(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load flag parser: %v", err)
	}

	return &FlagManager{
		parser: parser,
	}, nil
}

// GetFlag retrieves a flag value by its label
func (fm *FlagManager) GetFlag(label string) (uint32, error) {
	return fm.parser.GetFlag(label)
}

// type Handlers struct {
// 	fs                 *FSData
// 	db                 *gdbm.Database
// 	flagManager        *FlagManager
// 	accountFileHandler utils.AccountFileHandlerInterface
// 	accountService     server.AccountServiceInterface
// }

type Handlers struct {
	st                 *state.State
	ca                 cache.Memory
	userdataStore      db.Db
	flagManager        *asm.FlagParser
	accountFileHandler *utils.AccountFileHandler
	accountService     server.AccountServiceInterface
}

func NewHandlers(appFlags *asm.FlagParser, pe *persist.Persister, userdataStore db.Db) (*Handlers, error) {
	h := &Handlers{
		st:                 pe.State,
		ca:                 pe.GetMemory(),
		userdataStore:      userdataStore,
		flagManager:        appFlags,
		accountFileHandler: utils.NewAccountFileHandler(userdataStore),
		accountService:     &server.AccountService{},
	}
	if h.st == nil || h.ca == nil || h.userdataStore == nil || h.flagManager == nil {
		//logg.Errorf("have nil for essential value in handler", "state", h.st, "cache", h.ca, "store", h.userdataStore, "flags", h.flagManager)
		return nil, fmt.Errorf("have nil for essential value")
	}
	return h, nil
}

// Define the regex pattern as a constant
const pinPattern = `^\d{4}$`

// isValidPIN checks whether the given input is a 4 digit number
func isValidPIN(pin string) bool {
	match, _ := regexp.MatchString(pinPattern, pin)
	return match
}

// SetLanguage sets the language across the menu
func (h *Handlers) SetLanguage(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	inputStr := string(input)
	switch inputStr {
	case "0":
		res.FlagSet = []uint32{state.FLAG_LANG}
		res.Content = "eng"
	case "1":
		res.FlagSet = []uint32{state.FLAG_LANG}
		res.Content = "swa"
	default:
	}

	languageSetFlag, err := h.flagManager.GetFlag("flag_language_set")
	if err != nil {
		return res, err
	}
	res.FlagSet = append(res.FlagSet, languageSetFlag)

	return res, nil
}

func (h *Handlers) createAccountNoExist(ctx context.Context, sessionId string, res *resource.Result) error {
	accountResp, err := h.accountService.CreateAccount()
	data := map[utils.DataTyp]string{
		utils.DATA_TRACKING_ID:  accountResp.Result.TrackingId,
		utils.DATA_PUBLIC_KEY:   accountResp.Result.PublicKey,
		utils.DATA_CUSTODIAL_ID: accountResp.Result.CustodialId.String(),
	}

	for key, value := range data {
		err := utils.WriteEntry(ctx, h.userdataStore, sessionId, key, []byte(value))
		if err != nil {
			return err
		}
	}
	flag_account_created, _ := h.flagManager.GetFlag("flag_account_created")
	res.FlagSet = append(res.FlagSet, flag_account_created)
	return err

}

// CreateAccount checks if any account exists on the JSON data file, and if not
// creates an account on the API,
// sets the default values and flags
func (h *Handlers) CreateAccount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	_, err = utils.ReadEntry(ctx, h.userdataStore, sessionId, utils.DATA_ACCOUNT_CREATED)
	if err != nil {
		if db.IsNotFound(err) {
			fmt.Println("Creating an account because it doesn't exist")
			err = h.createAccountNoExist(ctx, sessionId, &res)
			if err != nil {
				return res, err
			}
		} else {
			fmt.Println("Error here:", err)
			err = h.createAccountNoExist(ctx, sessionId, &res)
			if err != nil {
				return res, err
			}
		}
	}
	return res, nil
}

// SavePin persists the user's PIN choice into the filesystem
func (h *Handlers) SavePin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_incorrect_pin, _ := h.flagManager.GetFlag("flag_incorrect_pin")

	accountPIN := string(input)
	// Validate that the PIN is a 4-digit number
	if !isValidPIN(accountPIN) {
		res.FlagSet = append(res.FlagSet, flag_incorrect_pin)
		return res, nil
	}

	res.FlagReset = append(res.FlagReset, flag_incorrect_pin)

	// key := []byte(AccountPin)
	// value := []byte(accountPIN)

	//h.db.Store(key, value, true)
	return res, nil
}

// SetResetSingleEdit sets and resets  flags to allow gradual editing of profile information.
func (h *Handlers) SetResetSingleEdit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	menuOption := string(input)

	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	flag_single_edit, _ := h.flagManager.GetFlag("flag_single_edit")

	switch menuOption {
	case "2":
		res.FlagReset = append(res.FlagReset, flag_allow_update)
		res.FlagSet = append(res.FlagSet, flag_single_edit)
	case "3":
		res.FlagReset = append(res.FlagReset, flag_allow_update)
		res.FlagSet = append(res.FlagSet, flag_single_edit)
	case "4":
		res.FlagReset = append(res.FlagReset, flag_allow_update)
		res.FlagSet = append(res.FlagSet, flag_single_edit)
	default:
		res.FlagReset = append(res.FlagReset, flag_single_edit)
	}

	return res, nil
}

// VerifyPin checks whether the confirmation PIN is similar to the account PIN
// If similar, it sets the USERFLAG_PIN_SET flag allowing the user
// to access the main menu
func (h *Handlers) VerifyPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_valid_pin, _ := h.flagManager.GetFlag("flag_valid_pin")
	flag_pin_mismatch, _ := h.flagManager.GetFlag("flag_pin_mismatch")
	flag_pin_set, _ := h.flagManager.GetFlag("flag_pin_set")

	// AccountPin, err := h.db.Fetch([]byte(AccountPin))
	// if err != nil {
	// 	return res, err
	// }
	AccountPin := []byte("2768")
	if bytes.Equal(input, AccountPin) {
		res.FlagSet = []uint32{flag_valid_pin}
		res.FlagReset = []uint32{flag_pin_mismatch}
		res.FlagSet = append(res.FlagSet, flag_pin_set)
	} else {
		res.FlagSet = []uint32{flag_pin_mismatch}
	}

	return res, nil
}

// codeFromCtx retrieves language codes from the context that can be used for handling translations
func codeFromCtx(ctx context.Context) string {
	var code string
	if ctx.Value("Language") != nil {
		lang := ctx.Value("Language").(lang.Language)
		code = lang.Code
	}
	return code
}

// SaveFirstname updates the first name in a JSON data file with the provided input.
func (h *Handlers) SaveFirstname(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	if len(input) > 0 {
		//name := string(input)
		//key := []byte(FirstName)
		//value := []byte(name)
		//h.db.Store(key, value, true)
	}

	return res, nil
}

// SaveFamilyname updates the family name in a JSON data file with the provided input.
func (h *Handlers) SaveFamilyname(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	if len(input) > 0 {
		//secondname := string(input)
		//key := []byte(FamilyName)
		//value := []byte(secondname)
		//h.db.Store(key, value, true)
	}

	return res, nil
}

// SaveYOB updates the Year of Birth(YOB) in a JSON data file with the provided input.
func (h *Handlers) SaveYob(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	yob := string(input)
	if len(yob) == 4 {
		//yob := string(input)
		//key := []byte(YearOfBirth)
		//value := []byte(yob)
		//h.db.Store(key, value, true)
	}

	return res, nil
}

// SaveLocation updates the location in a JSON data file with the provided input.
func (h *Handlers) SaveLocation(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	if len(input) > 0 {
		//location := string(input)
		//key := []byte(Location)
		//value := []byte(location)

		//h.db.Store(key, value, true)
	}

	return res, nil
}

// SaveGender updates the gender in a JSON data file with the provided input.
func (h *Handlers) SaveGender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	if len(input) > 0 {
		gender := string(input)
		switch gender {
		case "1":
			gender = "Male"
		case "2":
			gender = "Female"
		case "3":
			gender = "Unspecified"
		}
		//key := []byte(Gender)
		//value := []byte(gender)
		//h.db.Store(key, value, true)
	}
	return res, nil
}

// SaveOfferings updates the offerings(goods and services provided by the user) in a JSON data file with the provided input.
func (h *Handlers) SaveOfferings(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	if len(input) > 0 {
		//offerings := string(input)
		//key := []byte(Offerings)
		//value := []byte(offerings)
		//h.db.Store(key, value, true)
	}
	return res, nil
}

// ResetAllowUpdate resets the allowupdate flag that allows a user to update  profile data.
func (h *Handlers) ResetAllowUpdate(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")

	res.FlagReset = append(res.FlagReset, flag_allow_update)
	return res, nil
}

// ResetAccountAuthorized resets the account authorization flag after a successful PIN entry.
func (h *Handlers) ResetAccountAuthorized(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")

	res.FlagReset = append(res.FlagReset, flag_account_authorized)
	return res, nil
}

// CheckIdentifier retrieves the PublicKey from the JSON data file.
func (h *Handlers) CheckIdentifier(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	//publicKey, err := h.db.Fetch([]byte(PublicKeyKey))
	// if err != nil {
	// 	return res, err
	// }
	res.Content = "string(publicKey)"
	return res, nil
}

// Authorize attempts to unlock the next sequential nodes by verifying the provided PIN against the already set PIN.
// It sets the required flags that control the flow.
func (h *Handlers) Authorize(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	// flag_incorrect_pin, _ := h.flagManager.GetFlag("flag_incorrect_pin")
	// flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")
	// flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")

	// storedpin, err := h.db.Fetch([]byte(AccountPin))
	// if err == nil {
	// 	if len(input) == 4 {
	// 		if bytes.Equal(input, storedpin) {
	// 			if h.fs.St.MatchFlag(flag_account_authorized, false) {
	// 				res.FlagReset = append(res.FlagReset, flag_incorrect_pin)
	// 				res.FlagSet = append(res.FlagSet, flag_allow_update, flag_account_authorized)
	// 			} else {
	// 				res.FlagSet = append(res.FlagSet, flag_allow_update)
	// 				res.FlagReset = append(res.FlagReset, flag_account_authorized)
	// 			}
	// 		} else {
	// 			res.FlagSet = append(res.FlagSet, flag_incorrect_pin)
	// 			res.FlagReset = append(res.FlagReset, flag_account_authorized)
	// 			return res, nil
	// 		}
	// 	}
	// } else if errors.Is(err, gdbm.ErrItemNotFound) {
	// 	return res, err
	// } else {
	// 	return res, err
	// }
	return res, nil
}

// ResetIncorrectPin resets the incorrect pin flag  after a new PIN attempt.
func (h *Handlers) ResetIncorrectPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_incorrect_pin, _ := h.flagManager.GetFlag("flag_incorrect_pin")

	res.FlagReset = append(res.FlagReset, flag_incorrect_pin)
	return res, nil
}

// CheckAccountStatus queries the API using the TrackingId and sets flags
// based on the account status
func (h *Handlers) CheckAccountStatus(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_account_success, _ := h.flagManager.GetFlag("flag_account_success")
	flag_account_pending, _ := h.flagManager.GetFlag("flag_account_pending")

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	trackingId, err := utils.ReadEntry(ctx, h.userdataStore, sessionId, utils.DATA_TRACKING_ID)
	fmt.Println("Checking status with tracking id:", string(trackingId))

	status, err := h.accountService.CheckAccountStatus(string("1234"))

	if err != nil {
		fmt.Println("Error checking account status:", err)
		return res, err

	}

	// err = h.db.Store(toBytes(AccountStatus), toBytes(status), true)
	// if err != nil {
	// 	return res, nil
	// }

	// err = h.db.Store(toBytes(TrackingIdKey), toBytes(status), true)
	// if err != nil {
	// 	return res, nil
	// }

	if status == "SUCCESS" {
		res.FlagSet = append(res.FlagSet, flag_account_success)
		res.FlagReset = append(res.FlagReset, flag_account_pending)
	} else {
		res.FlagReset = append(res.FlagReset, flag_account_success)
		res.FlagSet = append(res.FlagSet, flag_account_pending)
	}
	return res, nil
}

// Quit displays the Thank you message and exits the menu
func (h *Handlers) Quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")

	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")

	res.Content = l.Get("Thank you for using Sarafu. Goodbye!")
	res.FlagReset = append(res.FlagReset, flag_account_authorized)
	return res, nil
}

// VerifyYob verifies the length of the given input
func (h *Handlers) VerifyYob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_incorrect_date_format, _ := h.flagManager.GetFlag("flag_incorrect_date_format")

	date := string(input)
	_, err := strconv.Atoi(date)
	if err != nil {
		// If conversion fails, input is not numeric
		res.FlagSet = append(res.FlagSet, flag_incorrect_date_format)
		return res, nil
	}

	if len(date) == 4 {
		res.FlagReset = append(res.FlagReset, flag_incorrect_date_format)
	} else {
		res.FlagSet = append(res.FlagSet, flag_incorrect_date_format)
	}

	return res, nil
}

// ResetIncorrectYob resets the incorrect date format flag after a new attempt
func (h *Handlers) ResetIncorrectYob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_incorrect_date_format, _ := h.flagManager.GetFlag("flag_incorrect_date_format")

	res.FlagReset = append(res.FlagReset, flag_incorrect_date_format)
	return res, nil
}

// CheckBalance retrieves the balance from the API using the "PublicKey" and sets
// the balance as the result content
func (h *Handlers) CheckBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	// publicKey, err := h.db.Fetch([]byte(PublicKeyKey))

	// if err != nil {
	// 	return res, err
	// }

	balance, err := h.accountService.CheckBalance(string("publicKey"))
	if err != nil {
		return res, nil
	}
	res.Content = balance

	return res, nil
}

// ValidateRecipient validates that the given input is a valid phone number.
func (h *Handlers) ValidateRecipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	recipient := string(input)

	flag_invalid_recipient, _ := h.flagManager.GetFlag("flag_invalid_recipient")

	if recipient != "0" {
		// mimic invalid number check
		if recipient == "000" {
			res.FlagSet = append(res.FlagSet, flag_invalid_recipient)
			res.Content = recipient

			return res, nil
		}

		// accountData["Recipient"] = recipient
		// key := []byte(Recipient)
		// value := []byte(recipient)

		// h.db.Store(key, value, true)
	}

	return res, nil
}

// TransactionReset resets the previous transaction data (Recipient and Amount)
// as well as the invalid flags
func (h *Handlers) TransactionReset(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_invalid_recipient, _ := h.flagManager.GetFlag("flag_invalid_recipient")
	flag_invalid_recipient_with_invite, _ := h.flagManager.GetFlag("flag_invalid_recipient_with_invite")

	// err := h.db.Delete([]byte(Amount))
	// if err != nil && !errors.Is(err, gdbm.ErrItemNotFound) {
	// 	return res, err
	// }
	// err = h.db.Delete([]byte(Recipient))
	// if err != nil && !errors.Is(err, gdbm.ErrItemNotFound) {
	// 	return res, err
	// }

	res.FlagReset = append(res.FlagReset, flag_invalid_recipient, flag_invalid_recipient_with_invite)

	return res, nil
}

// ResetTransactionAmount resets the transaction amount and invalid flag
func (h *Handlers) ResetTransactionAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_invalid_amount, _ := h.flagManager.GetFlag("flag_invalid_amount")

	// err := h.db.Delete([]byte(Amount))
	// if err != nil && !errors.Is(err, gdbm.ErrItemNotFound) {
	// 	return res, err
	// }

	res.FlagReset = append(res.FlagReset, flag_invalid_amount)

	return res, nil
}

// MaxAmount gets the current balance from the API and sets it as
// the result content.
func (h *Handlers) MaxAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	// publicKey, err := h.db.Fetch([]byte(PublicKeyKey))
	// if err != nil {
	// 	return res, err
	// }

	balance, err := h.accountService.CheckBalance(string("publicKey"))
	if err != nil {
		return res, nil
	}

	res.Content = balance

	return res, nil
}

// ValidateAmount ensures that the given input is a valid amount and that
// it is not more than the current balance.
func (h *Handlers) ValidateAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_invalid_amount, _ := h.flagManager.GetFlag("flag_invalid_amount")

	amountStr := string(input)
	// publicKey, err := h.db.Fetch([]byte(PublicKeyKey))

	// if err != nil {
	// 	return res, err
	// }

	balanceStr, err := h.accountService.CheckBalance(string("publicKey"))

	if err != nil {
		return res, err
	}
	res.Content = balanceStr

	// Parse the balance
	balanceParts := strings.Split(balanceStr, " ")
	if len(balanceParts) != 2 {
		return res, fmt.Errorf("unexpected balance format: %s", balanceStr)
	}
	balanceValue, err := strconv.ParseFloat(balanceParts[0], 64)
	if err != nil {
		return res, fmt.Errorf("failed to parse balance: %v", err)
	}

	// Extract numeric part from input
	re := regexp.MustCompile(`^(\d+(\.\d+)?)\s*(?:CELO)?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(amountStr))
	if len(matches) < 2 {
		res.FlagSet = append(res.FlagSet, flag_invalid_amount)
		res.Content = amountStr
		return res, nil
	}

	inputAmount, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		res.FlagSet = append(res.FlagSet, flag_invalid_amount)
		res.Content = amountStr
		return res, nil
	}

	if inputAmount > balanceValue {
		res.FlagSet = append(res.FlagSet, flag_invalid_amount)
		res.Content = amountStr
		return res, nil
	}

	res.Content = fmt.Sprintf("%.3f", inputAmount) // Format to 3 decimal places
	// key := []byte(Amount)
	// value := []byte(res.Content)
	// h.db.Store(key, value, true)

	if err != nil {
		return res, err
	}

	return res, nil
}

// GetRecipient returns the transaction recipient from a JSON data file.
func (h *Handlers) GetRecipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	// recipient, err := h.db.Fetch([]byte(Recipient))
	// if err != nil {
	// 	return res, err
	// }

	res.Content = string("recipient")

	return res, nil
}

// GetSender retrieves the public key from the Gdbm Db
func (h *Handlers) GetSender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	// publicKey, err := h.db.Fetch([]byte(PublicKeyKey))
	// if err != nil {
	// 	return res, err
	// }

	res.Content = string("publicKey")

	return res, nil
}

// GetAmount retrieves the amount from teh Gdbm Db
func (h *Handlers) GetAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	// amount, err := h.db.Fetch([]byte(Amount))
	// if err != nil {
	// 	return res, err
	// }
	res.Content = string("amount")

	return res, nil
}

// QuickWithBalance retrieves the balance for a given public key from the custodial balance API endpoint before
// gracefully exiting the session.
func (h *Handlers) QuitWithBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")

	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")
	// publicKey, err := h.db.Fetch([]byte(PublicKeyKey))
	// if err != nil {
	// 	return res, err
	// }
	balance, err := h.accountService.CheckBalance(string("publicKey"))
	if err != nil {
		return res, nil
	}
	res.Content = l.Get("Your account balance is %s", balance)
	res.FlagReset = append(res.FlagReset, flag_account_authorized)
	return res, nil
}

// InitiateTransaction returns a confirmation and resets the transaction data
// on the JSON file.
func (h *Handlers) InitiateTransaction(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")
	// TODO
	// Use the amount, recipient and sender to call the API and initialize the transaction

	// publicKey, err := h.db.Fetch([]byte(PublicKeyKey))
	// if err != nil {
	// 	return res, err
	// }
	// amount, err := h.db.Fetch([]byte(Amount))
	// if err != nil {
	// 	return res, err
	// }
	// recipient, err := h.db.Fetch([]byte(Recipient))
	// if err != nil {
	// 	return res, err
	// }

	//res.Content = l.Get("Your request has been sent. %s will receive %s from %s.", string(recipient), string(amount), string(publicKey))

	account_authorized_flag, err := h.flagManager.GetFlag("flag_account_authorized")
	if err != nil {
		return res, err
	}

	res.FlagReset = append(res.FlagReset, account_authorized_flag)
	return res, nil
}

// GetProfileInfo retrieves and formats the profile information of a user from a Gdbm backed storage.
func (h *Handlers) GetProfileInfo(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	// Define default values
	defaultValue := "Not provided"
	name := defaultValue
	familyName := defaultValue
	yob := defaultValue
	gender := defaultValue
	location := defaultValue
	offerings := defaultValue

	// Fetch data using a map for better organization
	// dataKeys := map[string]*string{
	// 	FirstName:   &name,
	// 	FamilyName:  &familyName,
	// 	YearOfBirth: &yob,
	// 	Location:    &location,
	// 	Gender:      &gender,
	// 	Offerings:   &offerings,
	// }

	// Iterate over keys and fetch values
	//iter := h.db.Iterator()
	// next := h.db.Iterator()
	// //defer iter.Close() // Ensure the iterator is closed
	// for key, err := next(); err == nil; key, err = next() {
	// 	if valuePointer, ok := dataKeys[string(key)]; ok {
	// 		// value, fetchErr := h.db.Fetch(key)
	// 		// if fetchErr == nil {
	// 		// 	*valuePointer = string(value)
	// 		// }
	// 	}
	// }

	// Construct the full name
	if familyName != defaultValue {
		if name == defaultValue {
			name = familyName
		} else {
			name = name + " " + familyName
		}
	}

	// Calculate age from year of birth
	var age string
	if yob != defaultValue {
		yobInt, err := strconv.Atoi(yob)
		if err != nil {
			return res, fmt.Errorf("invalid year of birth: %v", err)
		}
		age = strconv.Itoa(utils.CalculateAgeWithYOB(yobInt))
	} else {
		age = defaultValue
	}

	// Format the result
	formattedData := fmt.Sprintf("Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n", name, gender, age, location, offerings)
	res.Content = formattedData
	return res, nil
}

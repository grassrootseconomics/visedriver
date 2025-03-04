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
	"github.com/grassrootseconomics/eth-custodial/pkg/api"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/common"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"git.grassecon.net/urdt/ussd/remote"
	"gopkg.in/leonelquinteros/gotext.v1"

	"git.grassecon.net/urdt/ussd/internal/storage"
)

var (
	logg           = logging.NewVanilla().WithDomain("ussdmenuhandler")
	scriptDir      = path.Join("services", "registration")
	translationDir = path.Join(scriptDir, "locale")
	okResponse     *api.OKResponse
	errResponse    *api.ErrResponse
)

// Define the regex patterns as  constants
const (
	phoneRegex = `(\(\d{3}\)\s?|\d{3}[-.\s]?)?\d{3}[-.\s]?\d{4}`
	pinPattern = `^\d{4}$`
)

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

type Handlers struct {
	pe             *persist.Persister
	st             *state.State
	ca             cache.Memory
	userdataStore  common.DataStore
	adminstore     *utils.AdminStore
	flagManager    *asm.FlagParser
	accountService remote.AccountServiceInterface
	prefixDb       storage.PrefixDb
}

func NewHandlers(appFlags *asm.FlagParser, userdataStore db.Db, adminstore *utils.AdminStore, accountService remote.AccountServiceInterface) (*Handlers, error) {
	if userdataStore == nil {
		return nil, fmt.Errorf("cannot create handler with nil userdata store")
	}
	userDb := &common.UserDataStore{
		Db: userdataStore,
	}
	// Instantiate the SubPrefixDb with "vouchers" prefix
	prefixDb := storage.NewSubPrefixDb(userdataStore, []byte("vouchers"))

	h := &Handlers{
		userdataStore:  userDb,
		flagManager:    appFlags,
		adminstore:     adminstore,
		accountService: accountService,
		prefixDb:       prefixDb,
	}
	return h, nil
}

// isValidPIN checks whether the given input is a 4 digit number
func isValidPIN(pin string) bool {
	match, _ := regexp.MatchString(pinPattern, pin)
	return match
}

func isValidPhoneNumber(phonenumber string) bool {
	match, _ := regexp.MatchString(phoneRegex, phonenumber)
	return match
}

func (h *Handlers) WithPersister(pe *persist.Persister) *Handlers {
	if h.pe != nil {
		panic("persister already set")
	}
	h.pe = pe
	return h
}

func (h *Handlers) Init(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result
	if h.pe == nil {
		logg.WarnCtxf(ctx, "handler init called before it is ready or more than once", "state", h.st, "cache", h.ca)
		return r, nil
	}

	h.st = h.pe.GetState()
	h.ca = h.pe.GetMemory()

	sessionId, _ := ctx.Value("SessionId").(string)
	flag_admin_privilege, _ := h.flagManager.GetFlag("flag_admin_privilege")

	isAdmin, _ := h.adminstore.IsAdmin(sessionId)

	if isAdmin {
		r.FlagSet = append(r.FlagSet, flag_admin_privilege)
	} else {
		r.FlagReset = append(r.FlagReset, flag_admin_privilege)
	}

	if h.st == nil || h.ca == nil {
		logg.ErrorCtxf(ctx, "perister fail in handler", "state", h.st, "cache", h.ca)
		return r, fmt.Errorf("cannot get state and memory for handler")
	}
	h.pe = nil

	logg.DebugCtxf(ctx, "handler has been initialized", "state", h.st, "cache", h.ca)

	return r, nil
}

// SetLanguage sets the language across the menu
func (h *Handlers) SetLanguage(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	symbol, _ := h.st.Where()
	code := strings.Split(symbol, "_")[1]

	if !utils.IsValidISO639(code) {
		return res, nil
	}
	res.FlagSet = append(res.FlagSet, state.FLAG_LANG)
	res.Content = code

	languageSetFlag, err := h.flagManager.GetFlag("flag_language_set")
	if err != nil {
		logg.ErrorCtxf(ctx, "Error setting the languageSetFlag", "error", err)
		return res, err
	}
	res.FlagSet = append(res.FlagSet, languageSetFlag)

	return res, nil
}

func (h *Handlers) createAccountNoExist(ctx context.Context, sessionId string, res *resource.Result) error {
	flag_account_created, _ := h.flagManager.GetFlag("flag_account_created")
	r, err := h.accountService.CreateAccount(ctx)
	if err != nil {
		return err
	}
	trackingId := r.TrackingId
	publicKey := r.PublicKey

	data := map[common.DataTyp]string{
		common.DATA_TRACKING_ID: trackingId,
		common.DATA_PUBLIC_KEY:  publicKey,
	}
	store := h.userdataStore
	for key, value := range data {
		err = store.WriteEntry(ctx, sessionId, key, []byte(value))
		if err != nil {
			return err
		}
	}
	publicKeyNormalized, err := common.NormalizeHex(publicKey)
	if err != nil {
		return err
	}
	err = store.WriteEntry(ctx, publicKeyNormalized, common.DATA_PUBLIC_KEY_REVERSE, []byte(sessionId))
	if err != nil {
		return err
	}
	res.FlagSet = append(res.FlagSet, flag_account_created)
	return nil
}

// CreateAccount checks if any account exists on the JSON data file, and if not
// creates an account on the API,
// sets the default values and flags
func (h *Handlers) CreateAccount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	_, err = store.ReadEntry(ctx, sessionId, common.DATA_ACCOUNT_CREATED)
	if err != nil {
		if db.IsNotFound(err) {
			logg.InfoCtxf(ctx, "Creating an account because it doesn't exist")
			err = h.createAccountNoExist(ctx, sessionId, &res)
			if err != nil {
				logg.ErrorCtxf(ctx, "failed on createAccountNoExist", "error", err)
				return res, err
			}
		}
	}

	return res, nil
}

func (h *Handlers) CheckPinMisMatch(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	flag_pin_mismatch, _ := h.flagManager.GetFlag("flag_pin_mismatch")
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	blockedNumber, err := store.ReadEntry(ctx, sessionId, common.DATA_BLOCKED_NUMBER)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read blockedNumber entry with", "key", common.DATA_BLOCKED_NUMBER, "error", err)
		return res, err
	}
	temporaryPin, err := store.ReadEntry(ctx, string(blockedNumber), common.DATA_TEMPORARY_VALUE)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read temporaryPin entry with", "key", common.DATA_TEMPORARY_VALUE, "error", err)
		return res, err
	}
	if bytes.Equal(temporaryPin, input) {
		res.FlagReset = append(res.FlagReset, flag_pin_mismatch)
	} else {
		res.FlagSet = append(res.FlagSet, flag_pin_mismatch)
	}
	return res, nil
}

func (h *Handlers) VerifyNewPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	_, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	flag_valid_pin, _ := h.flagManager.GetFlag("flag_valid_pin")
	pinInput := string(input)
	// Validate that the PIN is a 4-digit number
	if isValidPIN(pinInput) {
		res.FlagSet = append(res.FlagSet, flag_valid_pin)
	} else {
		res.FlagReset = append(res.FlagReset, flag_valid_pin)
	}

	return res, nil
}

// SaveTemporaryPin saves the valid PIN input to the DATA_TEMPORARY_VALUE
// during the account creation process
// and during the change PIN process
func (h *Handlers) SaveTemporaryPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_incorrect_pin, _ := h.flagManager.GetFlag("flag_incorrect_pin")
	accountPIN := string(input)

	// Validate that the PIN is a 4-digit number
	if !isValidPIN(accountPIN) {
		res.FlagSet = append(res.FlagSet, flag_incorrect_pin)
		return res, nil
	}
	res.FlagReset = append(res.FlagReset, flag_incorrect_pin)
	store := h.userdataStore
	err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(accountPIN))
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to write temporaryAccountPIN entry with", "key", common.DATA_TEMPORARY_VALUE, "value", accountPIN, "error", err)
		return res, err
	}

	return res, nil
}

func (h *Handlers) SaveOthersTemporaryPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	store := h.userdataStore
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	temporaryPin := string(input)
	blockedNumber, err := store.ReadEntry(ctx, sessionId, common.DATA_BLOCKED_NUMBER)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read blockedNumber entry with", "key", common.DATA_BLOCKED_NUMBER, "error", err)
		return res, err
	}

	err = store.WriteEntry(ctx, string(blockedNumber), common.DATA_TEMPORARY_VALUE, []byte(temporaryPin))
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to write temporaryPin entry with", "key", common.DATA_TEMPORARY_VALUE, "value", temporaryPin, "error", err)
		return res, err
	}

	return res, nil
}

func (h *Handlers) ConfirmPinChange(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	flag_pin_mismatch, _ := h.flagManager.GetFlag("flag_pin_mismatch")

	store := h.userdataStore
	temporaryPin, err := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read temporaryPin entry with", "key", common.DATA_TEMPORARY_VALUE, "error", err)
		return res, err
	}
	if bytes.Equal(temporaryPin, input) {
		res.FlagReset = append(res.FlagReset, flag_pin_mismatch)
	} else {
		res.FlagSet = append(res.FlagSet, flag_pin_mismatch)
	}
	err = store.WriteEntry(ctx, sessionId, common.DATA_ACCOUNT_PIN, []byte(temporaryPin))
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to write temporaryPin entry with", "key", common.DATA_ACCOUNT_PIN, "value", temporaryPin, "error", err)
		return res, err
	}
	return res, nil
}

// VerifyCreatePin checks whether the confirmation PIN is similar to the temporary PIN
// If similar, it sets the USERFLAG_PIN_SET flag and writes the account PIN allowing the user
// to access the main menu
func (h *Handlers) VerifyCreatePin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	flag_valid_pin, _ := h.flagManager.GetFlag("flag_valid_pin")
	flag_pin_mismatch, _ := h.flagManager.GetFlag("flag_pin_mismatch")
	flag_pin_set, _ := h.flagManager.GetFlag("flag_pin_set")

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	temporaryPin, err := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read temporaryPin entry with", "key", common.DATA_TEMPORARY_VALUE, "error", err)
		return res, err
	}
	if bytes.Equal(input, temporaryPin) {
		res.FlagSet = []uint32{flag_valid_pin}
		res.FlagReset = []uint32{flag_pin_mismatch}
		res.FlagSet = append(res.FlagSet, flag_pin_set)
	} else {
		res.FlagSet = []uint32{flag_pin_mismatch}
	}

	err = store.WriteEntry(ctx, sessionId, common.DATA_ACCOUNT_PIN, []byte(temporaryPin))
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to write temporaryPin entry with", "key", common.DATA_ACCOUNT_PIN, "value", temporaryPin, "error", err)
		return res, err
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

// SaveFirstname updates the first name in the gdbm with the provided input.
func (h *Handlers) SaveFirstname(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	firstName := string(input)
	store := h.userdataStore
	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	allowUpdate := h.st.MatchFlag(flag_allow_update, true)
	if allowUpdate {
		temporaryFirstName, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
		err = store.WriteEntry(ctx, sessionId, common.DATA_FIRST_NAME, []byte(temporaryFirstName))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write firstName entry with", "key", common.DATA_FIRST_NAME, "value", temporaryFirstName, "error", err)
			return res, err
		}
	} else {
		err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(firstName))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write temporaryFirstName entry with", "key", common.DATA_TEMPORARY_VALUE, "value", firstName, "error", err)
			return res, err
		}
	}

	return res, nil
}

// SaveFamilyname updates the family name in the gdbm with the provided input.
func (h *Handlers) SaveFamilyname(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	store := h.userdataStore
	familyName := string(input)

	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	allowUpdate := h.st.MatchFlag(flag_allow_update, true)

	if allowUpdate {
		temporaryFamilyName, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
		err = store.WriteEntry(ctx, sessionId, common.DATA_FAMILY_NAME, []byte(temporaryFamilyName))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write familyName entry with", "key", common.DATA_FAMILY_NAME, "value", temporaryFamilyName, "error", err)
			return res, err
		}
	} else {
		err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(familyName))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write temporaryFamilyName entry with", "key", common.DATA_TEMPORARY_VALUE, "value", familyName, "error", err)
			return res, err
		}
	}

	return res, nil
}

// SaveYOB updates the Year of Birth(YOB) in the gdbm with the provided input.
func (h *Handlers) SaveYob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	yob := string(input)
	store := h.userdataStore
	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	allowUpdate := h.st.MatchFlag(flag_allow_update, true)

	if allowUpdate {
		temporaryYob, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
		err = store.WriteEntry(ctx, sessionId, common.DATA_YOB, []byte(temporaryYob))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write yob entry with", "key", common.DATA_TEMPORARY_VALUE, "value", temporaryYob, "error", err)
			return res, err
		}
	} else {
		err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(yob))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write temporaryYob entry with", "key", common.DATA_TEMPORARY_VALUE, "value", yob, "error", err)
			return res, err
		}
	}

	return res, nil
}

// SaveLocation updates the location in the gdbm with the provided input.
func (h *Handlers) SaveLocation(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	location := string(input)
	store := h.userdataStore

	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	allowUpdate := h.st.MatchFlag(flag_allow_update, true)

	if allowUpdate {
		temporaryLocation, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
		err = store.WriteEntry(ctx, sessionId, common.DATA_LOCATION, []byte(temporaryLocation))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write location entry with", "key", common.DATA_LOCATION, "value", temporaryLocation, "error", err)
			return res, err
		}
	} else {
		err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(location))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write temporaryLocation entry with", "key", common.DATA_TEMPORARY_VALUE, "value", location, "error", err)
			return res, err
		}
	}

	return res, nil
}

// SaveGender updates the gender in the gdbm with the provided input.
func (h *Handlers) SaveGender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	symbol, _ := h.st.Where()
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	gender := strings.Split(symbol, "_")[1]
	store := h.userdataStore
	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	allowUpdate := h.st.MatchFlag(flag_allow_update, true)

	if allowUpdate {
		temporaryGender, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
		err = store.WriteEntry(ctx, sessionId, common.DATA_GENDER, []byte(temporaryGender))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write gender entry with", "key", common.DATA_GENDER, "value", gender, "error", err)
			return res, err
		}
	} else {
		err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(gender))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write temporaryGender entry with", "key", common.DATA_TEMPORARY_VALUE, "value", gender, "error", err)
			return res, err
		}
	}

	return res, nil
}

// SaveOfferings updates the offerings(goods and services provided by the user) in the gdbm with the provided input.
func (h *Handlers) SaveOfferings(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	offerings := string(input)
	store := h.userdataStore

	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	allowUpdate := h.st.MatchFlag(flag_allow_update, true)

	if allowUpdate {
		temporaryOfferings, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)
		err = store.WriteEntry(ctx, sessionId, common.DATA_OFFERINGS, []byte(temporaryOfferings))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write offerings entry with", "key", common.DATA_TEMPORARY_VALUE, "value", offerings, "error", err)
			return res, err
		}
	} else {
		err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(offerings))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write temporaryOfferings entry with", "key", common.DATA_TEMPORARY_VALUE, "value", offerings, "error", err)
			return res, err
		}
	}

	return res, nil
}

// ResetAllowUpdate resets the allowupdate flag that allows a user to update  profile data.
func (h *Handlers) ResetAllowUpdate(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")
	res.FlagReset = append(res.FlagReset, flag_allow_update)
	return res, nil
}

// ResetAllowUpdate resets the allowupdate flag that allows a user to update  profile data.
func (h *Handlers) ResetValidPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	flag_valid_pin, _ := h.flagManager.GetFlag("flag_valid_pin")
	res.FlagReset = append(res.FlagReset, flag_valid_pin)
	return res, nil
}

// ResetAccountAuthorized resets the account authorization flag after a successful PIN entry.
func (h *Handlers) ResetAccountAuthorized(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")
	res.FlagReset = append(res.FlagReset, flag_account_authorized)
	return res, nil
}

// CheckIdentifier retrieves the PublicKey from the JSON data file.
func (h *Handlers) CheckIdentifier(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	publicKey, _ := store.ReadEntry(ctx, sessionId, common.DATA_PUBLIC_KEY)

	res.Content = string(publicKey)

	return res, nil
}

// Authorize attempts to unlock the next sequential nodes by verifying the provided PIN against the already set PIN.
// It sets the required flags that control the flow.
func (h *Handlers) Authorize(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	flag_incorrect_pin, _ := h.flagManager.GetFlag("flag_incorrect_pin")
	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")
	flag_allow_update, _ := h.flagManager.GetFlag("flag_allow_update")

	store := h.userdataStore
	AccountPin, err := store.ReadEntry(ctx, sessionId, common.DATA_ACCOUNT_PIN)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read AccountPin entry with", "key", common.DATA_ACCOUNT_PIN, "error", err)
		return res, err
	}
	if len(input) == 4 {
		if bytes.Equal(input, AccountPin) {
			if h.st.MatchFlag(flag_account_authorized, false) {
				res.FlagReset = append(res.FlagReset, flag_incorrect_pin)
				res.FlagSet = append(res.FlagSet, flag_allow_update, flag_account_authorized)
			} else {
				res.FlagSet = append(res.FlagSet, flag_allow_update)
				res.FlagReset = append(res.FlagReset, flag_account_authorized)
			}
		} else {
			res.FlagSet = append(res.FlagSet, flag_incorrect_pin)
			res.FlagReset = append(res.FlagReset, flag_account_authorized)
			return res, nil
		}
	} else {
		return res, nil
	}
	return res, nil
}

// ResetIncorrectPin resets the incorrect pin flag  after a new PIN attempt.
func (h *Handlers) ResetIncorrectPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	flag_incorrect_pin, _ := h.flagManager.GetFlag("flag_incorrect_pin")
	res.FlagReset = append(res.FlagReset, flag_incorrect_pin)
	return res, nil
}

// CheckAccountStatus queries the API using the TrackingId and sets flags
// based on the account status
func (h *Handlers) CheckAccountStatus(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	flag_account_success, _ := h.flagManager.GetFlag("flag_account_success")
	flag_account_pending, _ := h.flagManager.GetFlag("flag_account_pending")
	flag_api_error, _ := h.flagManager.GetFlag("flag_api_call_error")

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	store := h.userdataStore
	publicKey, err := store.ReadEntry(ctx, sessionId, common.DATA_PUBLIC_KEY)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read publicKey entry with", "key", common.DATA_PUBLIC_KEY, "error", err)
		return res, err
	}

	r, err := h.accountService.TrackAccountStatus(ctx, string(publicKey))
	if err != nil {
		res.FlagSet = append(res.FlagSet, flag_api_error)
		logg.ErrorCtxf(ctx, "failed on TrackAccountStatus", "error", err)
		return res, err
	}

	res.FlagReset = append(res.FlagReset, flag_api_error)

	if r.Active {
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
	var res resource.Result

	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")

	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")

	res.Content = l.Get("Thank you for using Sarafu. Goodbye!")
	res.FlagReset = append(res.FlagReset, flag_account_authorized)
	return res, nil
}

// QuitWithHelp displays helpline information then exits the menu
func (h *Handlers) QuitWithHelp(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")

	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")

	res.Content = l.Get("For more help,please call: 0757628885")
	res.FlagReset = append(res.FlagReset, flag_account_authorized)
	return res, nil
}

// VerifyYob verifies the length of the given input
func (h *Handlers) VerifyYob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	flag_incorrect_date_format, _ := h.flagManager.GetFlag("flag_incorrect_date_format")
	date := string(input)
	_, err = strconv.Atoi(date)
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
	var res resource.Result

	flag_incorrect_date_format, _ := h.flagManager.GetFlag("flag_incorrect_date_format")
	res.FlagReset = append(res.FlagReset, flag_incorrect_date_format)
	return res, nil
}

// CheckBalance retrieves the balance of the active voucher and sets
// the balance as the result content
func (h *Handlers) CheckBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")

	store := h.userdataStore

	// get the active sym and active balance
	activeSym, err := store.ReadEntry(ctx, sessionId, common.DATA_ACTIVE_SYM)
	if err != nil {
		if db.IsNotFound(err) {
			balance := "0.00"
			res.Content = l.Get("Balance: %s\n", balance)
			return res, nil
		}

		logg.ErrorCtxf(ctx, "failed to read activeSym entry with", "key", common.DATA_ACTIVE_SYM, "error", err)
		return res, err
	}

	activeBal, err := store.ReadEntry(ctx, sessionId, common.DATA_ACTIVE_BAL)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read activeBal entry with", "key", common.DATA_ACTIVE_BAL, "error", err)
		return res, err
	}

	res.Content = l.Get("Balance: %s\n", fmt.Sprintf("%s %s", activeBal, activeSym))

	return res, nil
}

func (h *Handlers) FetchCommunityBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")
	//TODO:
	//Check if the address is a community account,if then,get the actual balance
	res.Content = l.Get("Community Balance: 0.00")
	return res, nil
}

func (h *Handlers) ResetOthersPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	store := h.userdataStore
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	blockedPhonenumber, err := store.ReadEntry(ctx, sessionId, common.DATA_BLOCKED_NUMBER)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read blockedPhonenumber entry with", "key", common.DATA_BLOCKED_NUMBER, "error", err)
		return res, err
	}
	temporaryPin, err := store.ReadEntry(ctx, string(blockedPhonenumber), common.DATA_TEMPORARY_VALUE)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read temporaryPin entry with", "key", common.DATA_TEMPORARY_VALUE, "error", err)
		return res, err
	}
	err = store.WriteEntry(ctx, string(blockedPhonenumber), common.DATA_ACCOUNT_PIN, []byte(temporaryPin))
	if err != nil {
		return res, nil
	}

	return res, nil
}

func (h *Handlers) ResetUnregisteredNumber(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	flag_unregistered_number, _ := h.flagManager.GetFlag("flag_unregistered_number")
	res.FlagReset = append(res.FlagReset, flag_unregistered_number)
	return res, nil
}

func (h *Handlers) ValidateBlockedNumber(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	flag_unregistered_number, _ := h.flagManager.GetFlag("flag_unregistered_number")
	store := h.userdataStore
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	blockedNumber := string(input)
	_, err = store.ReadEntry(ctx, blockedNumber, common.DATA_PUBLIC_KEY)
	if !isValidPhoneNumber(blockedNumber) {
		res.FlagSet = append(res.FlagSet, flag_unregistered_number)
		return res, nil
	}
	if err != nil {
		if db.IsNotFound(err) {
			logg.InfoCtxf(ctx, "Invalid or unregistered number")
			res.FlagSet = append(res.FlagSet, flag_unregistered_number)
			return res, nil
		} else {
			logg.ErrorCtxf(ctx, "Error on ValidateBlockedNumber", "error", err)
			return res, err
		}
	}
	err = store.WriteEntry(ctx, sessionId, common.DATA_BLOCKED_NUMBER, []byte(blockedNumber))
	if err != nil {
		return res, nil
	}
	return res, nil
}

// ValidateRecipient validates that the given input is a valid phone number.
func (h *Handlers) ValidateRecipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	store := h.userdataStore

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	recipient := string(input)

	flag_invalid_recipient, _ := h.flagManager.GetFlag("flag_invalid_recipient")
	flag_invalid_recipient_with_invite, _ := h.flagManager.GetFlag("flag_invalid_recipient_with_invite")

	if recipient != "0" {
		if !isValidPhoneNumber(recipient) {
			res.FlagSet = append(res.FlagSet, flag_invalid_recipient)
			res.Content = recipient

			return res, nil
		}

		// save the recipient as the temporaryRecipient
		err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(recipient))
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write temporaryRecipient entry with", "key", common.DATA_TEMPORARY_VALUE, "value", recipient, "error", err)
			return res, err
		}

		publicKey, err := store.ReadEntry(ctx, recipient, common.DATA_PUBLIC_KEY)
		if err != nil {
			if db.IsNotFound(err) {
				logg.InfoCtxf(ctx, "Unregistered number")

				res.FlagSet = append(res.FlagSet, flag_invalid_recipient_with_invite)
				res.Content = recipient

				return res, nil
			}

			logg.ErrorCtxf(ctx, "failed to read publicKey entry with", "key", common.DATA_PUBLIC_KEY, "error", err)
			return res, err
		}

		err = store.WriteEntry(ctx, sessionId, common.DATA_RECIPIENT, publicKey)
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to write recipient entry with", "key", common.DATA_RECIPIENT, "value", string(publicKey), "error", err)
			return res, nil
		}
	}

	return res, nil
}

// TransactionReset resets the previous transaction data (Recipient and Amount)
// as well as the invalid flags
func (h *Handlers) TransactionReset(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_invalid_recipient, _ := h.flagManager.GetFlag("flag_invalid_recipient")
	flag_invalid_recipient_with_invite, _ := h.flagManager.GetFlag("flag_invalid_recipient_with_invite")
	store := h.userdataStore
	err = store.WriteEntry(ctx, sessionId, common.DATA_AMOUNT, []byte(""))
	if err != nil {
		return res, nil
	}

	err = store.WriteEntry(ctx, sessionId, common.DATA_RECIPIENT, []byte(""))
	if err != nil {
		return res, nil
	}

	res.FlagReset = append(res.FlagReset, flag_invalid_recipient, flag_invalid_recipient_with_invite)

	return res, nil
}

// InviteValidRecipient sends an invitation to the valid phone number.
func (h *Handlers) InviteValidRecipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	store := h.userdataStore

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")

	recipient, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)

	// TODO
	// send an invitation SMS
	// if successful
	// res.Content = l.Get("Your invitation to %s to join Sarafu Network has been sent.",  string(recipient))

	res.Content = l.Get("Your invite request for %s to Sarafu Network failed. Please try again later.", string(recipient))
	return res, nil
}

// ResetTransactionAmount resets the transaction amount and invalid flag
func (h *Handlers) ResetTransactionAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_invalid_amount, _ := h.flagManager.GetFlag("flag_invalid_amount")
	store := h.userdataStore
	err = store.WriteEntry(ctx, sessionId, common.DATA_AMOUNT, []byte(""))
	if err != nil {
		return res, nil
	}

	res.FlagReset = append(res.FlagReset, flag_invalid_amount)

	return res, nil
}

// MaxAmount gets the current balance from the API and sets it as
// the result content.
func (h *Handlers) MaxAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore

	activeBal, err := store.ReadEntry(ctx, sessionId, common.DATA_ACTIVE_BAL)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read activeBal entry with", "key", common.DATA_ACTIVE_BAL, "error", err)
		return res, err
	}

	res.Content = string(activeBal)

	return res, nil
}

// ValidateAmount ensures that the given input is a valid amount and that
// it is not more than the current balance.
func (h *Handlers) ValidateAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	flag_invalid_amount, _ := h.flagManager.GetFlag("flag_invalid_amount")
	store := h.userdataStore

	var balanceValue float64

	// retrieve the active balance
	activeBal, err := store.ReadEntry(ctx, sessionId, common.DATA_ACTIVE_BAL)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read activeBal entry with", "key", common.DATA_ACTIVE_BAL, "error", err)
		return res, err
	}
	balanceValue, err = strconv.ParseFloat(string(activeBal), 64)
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to convert the activeBal to a float", "error", err)
		return res, err
	}

	// Extract numeric part from the input amount
	amountStr := strings.TrimSpace(string(input))
	inputAmount, err := strconv.ParseFloat(amountStr, 64)
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

	// Format the amount with 2 decimal places before saving
	formattedAmount := fmt.Sprintf("%.2f", inputAmount)
	err = store.WriteEntry(ctx, sessionId, common.DATA_AMOUNT, []byte(formattedAmount))
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to write amount entry with", "key", common.DATA_AMOUNT, "value", formattedAmount, "error", err)
		return res, err
	}

	res.Content = formattedAmount
	return res, nil
}

// GetRecipient returns the transaction recipient phone number from the gdbm.
func (h *Handlers) GetRecipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	recipient, _ := store.ReadEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE)

	res.Content = string(recipient)

	return res, nil
}

// RetrieveBlockedNumber gets the current number during the pin reset for other's is in progress.
func (h *Handlers) RetrieveBlockedNumber(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	blockedNumber, _ := store.ReadEntry(ctx, sessionId, common.DATA_BLOCKED_NUMBER)

	res.Content = string(blockedNumber)

	return res, nil
}

// GetSender returns the sessionId (phoneNumber)
func (h *Handlers) GetSender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	res.Content = sessionId

	return res, nil
}

// GetAmount retrieves the amount from teh Gdbm Db
func (h *Handlers) GetAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore

	// retrieve the active symbol
	activeSym, err := store.ReadEntry(ctx, sessionId, common.DATA_ACTIVE_SYM)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read activeSym entry with", "key", common.DATA_ACTIVE_SYM, "error", err)
		return res, err
	}

	amount, _ := store.ReadEntry(ctx, sessionId, common.DATA_AMOUNT)

	res.Content = fmt.Sprintf("%s %s", string(amount), string(activeSym))

	return res, nil
}

// InitiateTransaction calls the TokenTransfer and returns a confirmation based on the result
func (h *Handlers) InitiateTransaction(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_account_authorized, _ := h.flagManager.GetFlag("flag_account_authorized")

	code := codeFromCtx(ctx)
	l := gotext.NewLocale(translationDir, code)
	l.AddDomain("default")

	data, err := common.ReadTransactionData(ctx, h.userdataStore, sessionId)
	if err != nil {
		return res, err
	}

	finalAmountStr, err := common.ParseAndScaleAmount(data.Amount, data.ActiveDecimal)
	if err != nil {
		return res, err
	}

	// Call TokenTransfer
	r, err := h.accountService.TokenTransfer(ctx, finalAmountStr, data.PublicKey, data.Recipient, data.ActiveAddress)
	if err != nil {
		flag_api_error, _ := h.flagManager.GetFlag("flag_api_call_error")
		res.FlagSet = append(res.FlagSet, flag_api_error)
		res.Content = l.Get("Your request failed. Please try again later.")
		logg.ErrorCtxf(ctx, "failed on TokenTransfer", "error", err)
		return res, nil
	}

	trackingId := r.TrackingId
	logg.InfoCtxf(ctx, "TokenTransfer", "trackingId", trackingId)

	res.Content = l.Get(
		"Your request has been sent. %s will receive %s %s from %s.",
		data.TemporaryValue,
		data.Amount,
		data.ActiveSym,
		sessionId,
	)

	res.FlagReset = append(res.FlagReset, flag_account_authorized)
	return res, nil
}

func (h *Handlers) GetCurrentProfileInfo(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var profileInfo []byte
	var err error
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	sm, _ := h.st.Where()
	parts := strings.SplitN(sm, "_", 2)
	filename := parts[1]
	dbKeyStr := "DATA_" + strings.ToUpper(filename)
	dbKey, err := common.StringToDataTyp(dbKeyStr)

	if err != nil {
		return res, err
	}
	store := h.userdataStore

	switch dbKey {
	case common.DATA_FIRST_NAME:
		profileInfo, err = store.ReadEntry(ctx, sessionId, common.DATA_FIRST_NAME)
		if err != nil {
			if db.IsNotFound(err) {
				res.Content = "Not provided"
				break
			}
			logg.ErrorCtxf(ctx, "Failed to read first name entry with", "key", "error", common.DATA_FIRST_NAME, err)
			return res, err
		}
		res.Content = string(profileInfo)
	case common.DATA_FAMILY_NAME:
		profileInfo, err = store.ReadEntry(ctx, sessionId, common.DATA_FAMILY_NAME)
		if err != nil {
			if db.IsNotFound(err) {
				res.Content = "Not provided"
				break
			}
			logg.ErrorCtxf(ctx, "Failed to read family name entry with", "key", "error", common.DATA_FAMILY_NAME, err)
			return res, err
		}
		res.Content = string(profileInfo)

	case common.DATA_GENDER:
		profileInfo, err = store.ReadEntry(ctx, sessionId, common.DATA_GENDER)
		if err != nil {
			if db.IsNotFound(err) {
				res.Content = "Not provided"
				break
			}
			logg.ErrorCtxf(ctx, "Failed to read gender entry with", "key", "error", common.DATA_GENDER, err)
			return res, err
		}
		res.Content = string(profileInfo)
	case common.DATA_YOB:
		profileInfo, err = store.ReadEntry(ctx, sessionId, common.DATA_YOB)
		if err != nil {
			if db.IsNotFound(err) {
				res.Content = "Not provided"
				break
			}
			logg.ErrorCtxf(ctx, "Failed to read year of birth(yob) entry with", "key", "error", common.DATA_YOB, err)
			return res, err
		}
		res.Content = string(profileInfo)

	case common.DATA_LOCATION:
		profileInfo, err = store.ReadEntry(ctx, sessionId, common.DATA_LOCATION)
		if err != nil {
			if db.IsNotFound(err) {
				res.Content = "Not provided"
				break
			}
			logg.ErrorCtxf(ctx, "Failed to read location entry with", "key", "error", common.DATA_LOCATION, err)
			return res, err
		}
		res.Content = string(profileInfo)
	case common.DATA_OFFERINGS:
		profileInfo, err = store.ReadEntry(ctx, sessionId, common.DATA_OFFERINGS)
		if err != nil {
			if db.IsNotFound(err) {
				res.Content = "Not provided"
				break
			}
			logg.ErrorCtxf(ctx, "Failed to read offerings entry with", "key", "error", common.DATA_OFFERINGS, err)
			return res, err
		}
		res.Content = string(profileInfo)
	default:
		break
	}

	return res, nil
}

func (h *Handlers) GetProfileInfo(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var defaultValue string
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	language, ok := ctx.Value("Language").(lang.Language)
	if !ok {
		return res, fmt.Errorf("value for 'Language' is not of type lang.Language")
	}
	code := language.Code
	if code == "swa" {
		defaultValue = "Haipo"
	} else {
		defaultValue = "Not Provided"
	}

	// Helper function to handle nil byte slices and convert them to string
	getEntryOrDefault := func(entry []byte, err error) string {
		if err != nil || entry == nil {
			return defaultValue
		}
		return string(entry)
	}
	store := h.userdataStore
	// Retrieve user data as strings with fallback to defaultValue
	firstName := getEntryOrDefault(store.ReadEntry(ctx, sessionId, common.DATA_FIRST_NAME))
	familyName := getEntryOrDefault(store.ReadEntry(ctx, sessionId, common.DATA_FAMILY_NAME))
	yob := getEntryOrDefault(store.ReadEntry(ctx, sessionId, common.DATA_YOB))
	gender := getEntryOrDefault(store.ReadEntry(ctx, sessionId, common.DATA_GENDER))
	location := getEntryOrDefault(store.ReadEntry(ctx, sessionId, common.DATA_LOCATION))
	offerings := getEntryOrDefault(store.ReadEntry(ctx, sessionId, common.DATA_OFFERINGS))

	// Construct the full name
	name := defaultValue
	if familyName != defaultValue {
		if firstName == defaultValue {
			name = familyName
		} else {
			name = firstName + " " + familyName
		}
	}

	// Calculate age from year of birth
	age := defaultValue
	if yob != defaultValue {
		if yobInt, err := strconv.Atoi(yob); err == nil {
			age = strconv.Itoa(utils.CalculateAgeWithYOB(yobInt))
		} else {
			return res, fmt.Errorf("invalid year of birth: %v", err)
		}
	}
	switch language.Code {
	case "eng":
		res.Content = fmt.Sprintf(
			"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
			name, gender, age, location, offerings,
		)
	case "swa":
		res.Content = fmt.Sprintf(
			"Jina: %s\nJinsia: %s\nUmri: %s\nEneo: %s\nUnauza: %s\n",
			name, gender, age, location, offerings,
		)
	default:
		res.Content = fmt.Sprintf(
			"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
			name, gender, age, location, offerings,
		)
	}

	return res, nil
}

// SetDefaultVoucher retrieves the current vouchers
// and sets the first as the default voucher, if no active voucher is set
func (h *Handlers) SetDefaultVoucher(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	var err error
	store := h.userdataStore

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_no_active_voucher, _ := h.flagManager.GetFlag("flag_no_active_voucher")

	// check if the user has an active sym
	_, err = store.ReadEntry(ctx, sessionId, common.DATA_ACTIVE_SYM)

	if err != nil {
		if db.IsNotFound(err) {
			publicKey, err := store.ReadEntry(ctx, sessionId, common.DATA_PUBLIC_KEY)
			if err != nil {
				logg.ErrorCtxf(ctx, "failed to read publicKey entry with", "key", common.DATA_PUBLIC_KEY, "error", err)
				return res, err
			}

			// Fetch vouchers from the API using the public key
			vouchersResp, err := h.accountService.FetchVouchers(ctx, string(publicKey))
			if err != nil {
				res.FlagSet = append(res.FlagSet, flag_no_active_voucher)
				return res, nil
			}

			// Return if there is no voucher
			if len(vouchersResp) == 0 {
				res.FlagSet = append(res.FlagSet, flag_no_active_voucher)
				return res, nil
			}

			// Use only the first voucher
			firstVoucher := vouchersResp[0]
			defaultSym := firstVoucher.TokenSymbol
			defaultBal := firstVoucher.Balance
			defaultDec := firstVoucher.TokenDecimals
			defaultAddr := firstVoucher.ContractAddress

			// Scale down the balance
			scaledBalance := common.ScaleDownBalance(defaultBal, defaultDec)

			// set the active symbol
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_SYM, []byte(defaultSym))
			if err != nil {
				logg.ErrorCtxf(ctx, "failed to write defaultSym entry with", "key", common.DATA_ACTIVE_SYM, "value", defaultSym, "error", err)
				return res, err
			}
			// set the active balance
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_BAL, []byte(scaledBalance))
			if err != nil {
				logg.ErrorCtxf(ctx, "failed to write defaultBal entry with", "key", common.DATA_ACTIVE_BAL, "value", scaledBalance, "error", err)
				return res, err
			}
			// set the active decimals
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_DECIMAL, []byte(defaultDec))
			if err != nil {
				logg.ErrorCtxf(ctx, "failed to write defaultDec entry with", "key", common.DATA_ACTIVE_DECIMAL, "value", defaultDec, "error", err)
				return res, err
			}
			// set the active contract address
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_ADDRESS, []byte(defaultAddr))
			if err != nil {
				logg.ErrorCtxf(ctx, "failed to write defaultAddr entry with", "key", common.DATA_ACTIVE_ADDRESS, "value", defaultAddr, "error", err)
				return res, err
			}

			return res, nil
		}

		logg.ErrorCtxf(ctx, "failed to read activeSym entry with", "key", common.DATA_ACTIVE_SYM, "error", err)
		return res, err
	}

	res.FlagReset = append(res.FlagReset, flag_no_active_voucher)

	return res, nil
}

// CheckVouchers retrieves the token holdings from the API using the "PublicKey" and stores
// them to gdbm
func (h *Handlers) CheckVouchers(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	store := h.userdataStore
	publicKey, err := store.ReadEntry(ctx, sessionId, common.DATA_PUBLIC_KEY)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read publicKey entry with", "key", common.DATA_PUBLIC_KEY, "error", err)
		return res, err
	}

	// Fetch vouchers from the API using the public key
	vouchersResp, err := h.accountService.FetchVouchers(ctx, string(publicKey))
	if err != nil {
		return res, nil
	}

	data := common.ProcessVouchers(vouchersResp)

	// Store all voucher data
	dataMap := map[string]string{
		"sym":  data.Symbols,
		"bal":  data.Balances,
		"deci": data.Decimals,
		"addr": data.Addresses,
	}

	for key, value := range dataMap {
		if err := h.prefixDb.Put(ctx, []byte(key), []byte(value)); err != nil {
			return res, nil
		}
	}

	return res, nil
}

// GetVoucherList fetches the list of vouchers and formats them
func (h *Handlers) GetVoucherList(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	// Read vouchers from the store
	voucherData, err := h.prefixDb.Get(ctx, []byte("sym"))
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to read the voucherData from prefixDb", "error", err)
		return res, err
	}

	res.Content = string(voucherData)

	return res, nil
}

// ViewVoucher retrieves the token holding and balance from the subprefixDB
// and displays it to the user for them to select it
func (h *Handlers) ViewVoucher(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_incorrect_voucher, _ := h.flagManager.GetFlag("flag_incorrect_voucher")

	inputStr := string(input)
	if inputStr == "0" || inputStr == "99" {
		res.FlagReset = append(res.FlagReset, flag_incorrect_voucher)
		return res, nil
	}

	metadata, err := common.GetVoucherData(ctx, h.prefixDb, inputStr)
	if err != nil {
		return res, fmt.Errorf("failed to retrieve voucher data: %v", err)
	}

	if metadata == nil {
		res.FlagSet = append(res.FlagSet, flag_incorrect_voucher)
		return res, nil
	}

	if err := common.StoreTemporaryVoucher(ctx, h.userdataStore, sessionId, metadata); err != nil {
		logg.ErrorCtxf(ctx, "failed on StoreTemporaryVoucher", "error", err)
		return res, err
	}

	res.FlagReset = append(res.FlagReset, flag_incorrect_voucher)
	res.Content = fmt.Sprintf("%s\n%s", metadata.TokenSymbol, metadata.Balance)

	return res, nil
}

// SetVoucher retrieves the temp voucher data and sets it as the active data
func (h *Handlers) SetVoucher(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	// Get temporary data
	tempData, err := common.GetTemporaryVoucherData(ctx, h.userdataStore, sessionId)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed on GetTemporaryVoucherData", "error", err)
		return res, err
	}

	// Set as active and clear temporary data
	if err := common.UpdateVoucherData(ctx, h.userdataStore, sessionId, tempData); err != nil {
		logg.ErrorCtxf(ctx, "failed on UpdateVoucherData", "error", err)
		return res, err
	}

	res.Content = tempData.TokenSymbol
	return res, nil
}

// GetVoucherDetails retrieves the voucher details
func (h *Handlers) GetVoucherDetails(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	store := h.userdataStore
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_api_error, _ := h.flagManager.GetFlag("flag_api_call_error")

	// get the active address
	activeAddress, err := store.ReadEntry(ctx, sessionId, common.DATA_ACTIVE_ADDRESS)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read activeAddress entry with", "key", common.DATA_ACTIVE_ADDRESS, "error", err)
		return res, err
	}

	// use the voucher contract address to get the data from the API
	voucherData, err := h.accountService.VoucherData(ctx, string(activeAddress))
	if err != nil {
		res.FlagSet = append(res.FlagSet, flag_api_error)
		return res, nil
	}

	tokenSymbol := voucherData.TokenSymbol
	tokenName := voucherData.TokenName

	res.Content = fmt.Sprintf("%s %s", tokenSymbol, tokenName)

	return res, nil
}

// CheckTransactions retrieves the transactions from the API using the "PublicKey" and stores to prefixDb
func (h *Handlers) CheckTransactions(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}

	flag_no_transfers, _ := h.flagManager.GetFlag("flag_no_transfers")
	flag_api_error, _ := h.flagManager.GetFlag("flag_api_error")

	store := h.userdataStore
	publicKey, err := store.ReadEntry(ctx, sessionId, common.DATA_PUBLIC_KEY)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read publicKey entry with", "key", common.DATA_PUBLIC_KEY, "error", err)
		return res, err
	}

	// Fetch transactions from the API using the public key
	transactionsResp, err := h.accountService.FetchTransactions(ctx, string(publicKey))
	if err != nil {
		res.FlagSet = append(res.FlagSet, flag_api_error)
		logg.ErrorCtxf(ctx, "failed on FetchTransactions", "error", err)
		return res, err
	}

	// Return if there are no transactions
	if len(transactionsResp) == 0 {
		res.FlagSet = append(res.FlagSet, flag_no_transfers)
		return res, nil
	}

	data := common.ProcessTransfers(transactionsResp)

	// Store all transaction data
	dataMap := map[string]string{
		"txfrom": data.Senders,
		"txto":   data.Recipients,
		"txval":  data.TransferValues,
		"txaddr": data.Addresses,
		"txhash": data.TxHashes,
		"txdate": data.Dates,
		"txsym":  data.Symbols,
		"txdeci": data.Decimals,
	}

	for key, value := range dataMap {
		if err := h.prefixDb.Put(ctx, []byte(key), []byte(value)); err != nil {
			logg.ErrorCtxf(ctx, "failed to write to prefixDb", "error", err)
			return res, err
		}
	}

	res.FlagReset = append(res.FlagReset, flag_no_transfers)

	return res, nil
}

// GetTransactionsList fetches the list of transactions and formats them
func (h *Handlers) GetTransactionsList(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	publicKey, err := store.ReadEntry(ctx, sessionId, common.DATA_PUBLIC_KEY)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read publicKey entry with", "key", common.DATA_PUBLIC_KEY, "error", err)
		return res, err
	}

	// Read transactions from the store and format them
	TransactionSenders, err := h.prefixDb.Get(ctx, []byte("txfrom"))
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to read the TransactionSenders from prefixDb", "error", err)
		return res, err
	}
	TransactionSyms, err := h.prefixDb.Get(ctx, []byte("txsym"))
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to read the TransactionSyms from prefixDb", "error", err)
		return res, err
	}
	TransactionValues, err := h.prefixDb.Get(ctx, []byte("txval"))
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to read the TransactionValues from prefixDb", "error", err)
		return res, err
	}
	TransactionDates, err := h.prefixDb.Get(ctx, []byte("txdate"))
	if err != nil {
		logg.ErrorCtxf(ctx, "Failed to read the TransactionDates from prefixDb", "error", err)
		return res, err
	}

	// Parse the data
	senders := strings.Split(string(TransactionSenders), "\n")
	syms := strings.Split(string(TransactionSyms), "\n")
	values := strings.Split(string(TransactionValues), "\n")
	dates := strings.Split(string(TransactionDates), "\n")

	var formattedTransactions []string
	for i := 0; i < len(senders); i++ {
		sender := strings.TrimSpace(senders[i])
		sym := strings.TrimSpace(syms[i])
		value := strings.TrimSpace(values[i])
		date := strings.Split(strings.TrimSpace(dates[i]), " ")[0]

		status := "received"
		if sender == string(publicKey) {
			status = "sent"
		}

		formattedTransactions = append(formattedTransactions, fmt.Sprintf("%d:%s %s %s %s", i+1, status, value, sym, date))
	}

	res.Content = strings.Join(formattedTransactions, "\n")

	return res, nil
}

// ViewTransactionStatement retrieves the transaction statement
// and displays it to the user
func (h *Handlers) ViewTransactionStatement(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var res resource.Result
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return res, fmt.Errorf("missing session")
	}
	store := h.userdataStore
	publicKey, err := store.ReadEntry(ctx, sessionId, common.DATA_PUBLIC_KEY)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to read publicKey entry with", "key", common.DATA_PUBLIC_KEY, "error", err)
		return res, err
	}

	flag_incorrect_statement, _ := h.flagManager.GetFlag("flag_incorrect_statement")

	inputStr := string(input)
	if inputStr == "0" || inputStr == "99" || inputStr == "11" || inputStr == "22" {
		res.FlagReset = append(res.FlagReset, flag_incorrect_statement)
		return res, nil
	}

	// Convert input string to integer
	index, err := strconv.Atoi(strings.TrimSpace(inputStr))
	if err != nil {
		return res, fmt.Errorf("invalid input: must be a number between 1 and 10")
	}

	if index < 1 || index > 10 {
		return res, fmt.Errorf("invalid input: index must be between 1 and 10")
	}

	statement, err := common.GetTransferData(ctx, h.prefixDb, string(publicKey), index)
	if err != nil {
		return res, fmt.Errorf("failed to retrieve transfer data: %v", err)
	}

	if statement == "" {
		res.FlagSet = append(res.FlagSet, flag_incorrect_statement)
		return res, nil
	}

	res.FlagReset = append(res.FlagReset, flag_incorrect_statement)
	res.Content = statement

	return res, nil
}

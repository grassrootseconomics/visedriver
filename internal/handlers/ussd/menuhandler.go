package ussd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/handlers/server"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/utils"
)

type FSData struct {
	Path string
	St   *state.State
}

type Handlers struct {
	fs *FSData
}

func NewHandlers(path string, st *state.State) *Handlers {
	return &Handlers{
		fs: &FSData{
			Path: path,
			St:   st,
		},
	}
}

func (h *Handlers) SetLanguage(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	inputStr := string(input)
	res := resource.Result{}
	switch inputStr {
	case "0":
		res.FlagSet = []uint32{state.FLAG_LANG}
		res.Content = "eng"
	case "1":
		res.FlagSet = []uint32{state.FLAG_LANG}
		res.Content = "swa"
	default:
	}

	res.FlagSet = append(res.FlagSet, models.USERFLAG_LANGUAGE_SET)

	return res, nil
}

func (h *Handlers) CreateAccount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return res, err
	}
	f.Close()
	accountResp, err := server.CreateAccount()
	if err != nil {
		fmt.Println("Failed to create account:", err)
		return res, err
	}

	accountData := map[string]string{
		"TrackingId":  accountResp.Result.TrackingId,
		"PublicKey":   accountResp.Result.PublicKey,
		"CustodialId": accountResp.Result.CustodialId.String(),
		"Status":      "PENDING",
	}

	jsonData, err := json.Marshal(accountData)
	if err != nil {
		return res, err
	}

	err = os.WriteFile(fp, jsonData, 0644)
	if err != nil {
		return res, err
	}
	res.FlagSet = append(res.FlagSet, models.USERFLAG_ACCOUNT_CREATED)
	return res, err
}

func (h *Handlers) SavePin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	accountPIN := string(input)
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	accountData["AccountPIN"] = accountPIN

	updatedJsonData, err := json.Marshal(accountData)
	if err != nil {
		return res, err
	}

	err = os.WriteFile(fp, updatedJsonData, 0644)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (h *Handlers) VerifyPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	if bytes.Equal(input, []byte(accountData["AccountPIN"])) {
		res.FlagSet = []uint32{models.USERFLAG_VALIDPIN}
		res.FlagReset = []uint32{models.USERFLAG_PINMISMATCH}
	} else {
		res.FlagSet = []uint32{models.USERFLAG_PINMISMATCH}
	}

	return res, nil
}

func codeFromCtx(ctx context.Context) string {
	var code string
	engine.Logg.DebugCtxf(ctx, "in msg", "ctx", ctx, "val", code)
	if ctx.Value("Language") != nil {
		lang := ctx.Value("Language").(lang.Language)
		code = lang.Code
	}
	return code
}

func (h *Handlers) SaveFirstname(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	if len(input) > 0 {
		name := string(input)
		accountData["FirstName"] = name
		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}

		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

func (h *Handlers) SaveFamilyname(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	if len(input) > 0 {
		//Save name
		secondname := string(input)
		accountData["FamilyName"] = secondname
		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}

		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

func (h *Handlers) SaveYob(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"
	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}
	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	yob := string(input)
	if len(yob) > 4 {
		yob := string(input)
		accountData["YOB"] = yob
		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}

		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

func (h *Handlers) SaveLocation(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"
	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}
	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	if len(input) > 0 {
		location := string(input)
		accountData["Location"] = location
		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}

		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}

	}

	return res, nil
}

func (h *Handlers) SaveGender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"
	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}
	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	if len(input) > 0 {
		gender := string(input)

		switch gender {
		case "1":
			gender = "Male"
		case "2":
			gender = "Female"
		case "3":
			gender = "Other"
		}
		accountData["Gender"] = gender
		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}

		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func (h *Handlers) SaveOfferings(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"
	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}
	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	if len(input) > 0 {
		offerings := string(input)
		accountData["Offerings"] = offerings
		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}
		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func (h *Handlers) ResetUnlockForUpdate(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_UNLOCKFORUPDATE)
	return res, nil
}

func (h *Handlers) ResetAccountUnlocked(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
	return res, nil
}

func (h *Handlers) CheckIdentifier(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	res.Content = accountData["PublicKey"]

	return res, nil
}

func (h *Handlers) Unlock(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	pin := string(input)
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	if len(input) > 1 {
		if pin != accountData["AccountPIN"] {
			res.FlagSet = append(res.FlagSet, models.USERFLAG_INCORRECTPIN)
			res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
			return res, nil
		}
		if h.fs.St.MatchFlag(models.USERFLAG_ACCOUNT_UNLOCKED, false) {
			res.FlagReset = append(res.FlagReset, models.USERFLAG_INCORRECTPIN)
			res.FlagSet = append(res.FlagSet, models.USERFLAG_UNLOCKFORUPDATE)
			res.FlagSet = append(res.FlagSet, models.USERFLAG_ACCOUNT_UNLOCKED)
		} else {
			res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
		}
	}
	return res, nil
}

func (h *Handlers) ResetIncorrectPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_INCORRECTPIN)
	return res, nil
}

func (h *Handlers) CheckAccountStatus(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	status, err := server.CheckAccountStatus(accountData["TrackingId"])

	if err != nil {
		fmt.Println("Error checking account status:", err)
		return res, nil
	}

	accountData["Status"] = status

	if status == "SUCCESS" {
		res.FlagSet = append(res.FlagSet, models.USERFLAG_ACCOUNT_SUCCESS)
		res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_PENDING)
	} else {
		res.FlagReset = append(res.FlagSet, models.USERFLAG_ACCOUNT_SUCCESS)
		res.FlagSet = append(res.FlagReset, models.USERFLAG_ACCOUNT_PENDING)
	}

	updatedJsonData, err := json.Marshal(accountData)
	if err != nil {
		return res, err
	}

	err = os.WriteFile(fp, updatedJsonData, 0644)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (h *Handlers) Quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	switch codeFromCtx(ctx) {
	case "swa":
		res.Content = "Asante kwa kutumia huduma ya Sarafu. Kwaheri!"
	default:
		res.Content = "Thank you for using Sarafu. Goodbye!"
	}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
	return res, nil
}

func (h *Handlers) VerifyYob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	date := string(input)

	dateRegex := regexp.MustCompile(`^\d{2}/\d{2}/\d{4}$`)
	isCorrectFormat := dateRegex.MatchString(date)
	if !isCorrectFormat {
		res.FlagSet = append(res.FlagSet, models.USERFLAG_INCORRECTDATEFORMAT)
	} else {
		res.FlagReset = append(res.FlagReset, models.USERFLAG_INCORRECTDATEFORMAT)
	}

	return res, nil
}

func (h *Handlers) ResetIncorrectYob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_INCORRECTDATEFORMAT)
	return res, nil
}

func (h *Handlers) CheckBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	fp := h.fs.Path + "_data"
	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}
	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	balance, err := server.CheckBalance(accountData["PublicKey"])
	if err != nil {
		return res, nil
	}
	res.Content = balance

	return res, nil
}

func (h *Handlers) ValidateRecipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	recipient := string(input)

	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	if recipient != "0" {
		// mimic invalid number check
		if recipient == "000" {
			res.FlagSet = append(res.FlagSet, models.USERFLAG_INVALID_RECIPIENT)
			res.Content = recipient

			return res, nil
		}

		accountData["Recipient"] = recipient

		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}

		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

func (h *Handlers) TransactionReset(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	// reset the recipient
	accountData["Recipient"] = ""

	updatedJsonData, err := json.Marshal(accountData)
	if err != nil {
		return res, err
	}

	err = os.WriteFile(fp, updatedJsonData, 0644)
	if err != nil {
		return res, err
	}

	res.FlagReset = append(res.FlagReset, models.USERFLAG_INVALID_RECIPIENT, models.USERFLAG_INVALID_RECIPIENT_WITH_INVITE)

	return res, nil
}

func (h *Handlers) ResetTransactionAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	// reset the amount
	accountData["Amount"] = ""

	updatedJsonData, err := json.Marshal(accountData)
	if err != nil {
		return res, err
	}

	err = os.WriteFile(fp, updatedJsonData, 0644)
	if err != nil {
		return res, err
	}

	res.FlagReset = append(res.FlagReset, models.USERFLAG_INVALID_AMOUNT)

	return res, nil
}

func (h *Handlers) MaxAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	// mimic a max amount
	res.Content = "10.00"

	return res, nil
}

func (h *Handlers) ValidateAmount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	amount := string(input)

	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	if amount != "0" {
		// mimic invalid amount
		if amount == "00" {
			res.FlagSet = append(res.FlagSet, models.USERFLAG_INVALID_AMOUNT)
			res.Content = amount

			return res, nil
		}

		res.Content = amount

		accountData["Amount"] = amount

		updatedJsonData, err := json.Marshal(accountData)
		if err != nil {
			return res, err
		}

		err = os.WriteFile(fp, updatedJsonData, 0644)
		if err != nil {
			return res, err
		}

		return res, nil
	}

	return res, nil
}

func (h *Handlers) GetRecipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	res.Content = accountData["Recipient"]

	return res, nil
}

func (h *Handlers) GetProfileInfo(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	name := accountData["FirstName"] + " " + accountData["FamilyName"]
	gender := accountData["Gender"]
	yob := accountData["YOB"]
	location := accountData["Location"]
	offerings := accountData["Offerings"]

	layout := "02/01/2006"

	birthdate, err := time.Parse(layout, yob)
	if err != nil {
		return res, err
	}

	currentDate := time.Now()
	formattedDate := currentDate.Format(layout)
	today, err := time.Parse(layout, formattedDate)
	if err != nil {
		return res, nil
	}
	age := utils.CalculateAge(birthdate, today)

	formattedData := fmt.Sprintf("Name: %s\nGender: %s\nAge: %d\nLocation: %s\nYou provide: %s\n", name, gender, age, location, offerings)
	res.Content = formattedData

	return res, nil
}

func (h *Handlers) GetSender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	res.Content = accountData["PublicKey"]

	return res, nil
}

func (h *Handlers) QuitWithBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := h.fs.Path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	balance, err := server.CheckBalance(accountData["PublicKey"])
	if err != nil {
		return res, nil
	}
	res.Content = fmt.Sprintf("Your account balance is: %s", balance)
	res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
	return res, nil
}

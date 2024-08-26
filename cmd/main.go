package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
	"time"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/server/handlers"
	"git.grassecon.net/urdt/ussd/internal/utils"
)

type fsData struct {
	path string
	st   *state.State
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

func (fsd *fsData) save_firstname(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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

func (fsd *fsData) save_familyname(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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
func (fsd *fsData) save_yob(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"
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

func (fsd *fsData) save_location(cxt context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"
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

func (fsd *fsData) save_gender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"
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

func (fsd *fsData) save_offerings(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"
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

func (fsd *fsData) set_language(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

func (fsd *fsData) create_account(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return res, err
	}
	f.Close()
	accountResp, err := handlers.CreateAccount()
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

func (fsd *fsData) reset_unlock_for_update(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_UNLOCKFORUPDATE)
	return res, nil
}

func (fsd *fsData) reset_account_unlocked(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
	return res, nil
}

func (fsd *fsData) check_identifier(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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

func (fsd *fsData) unlock(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	pin := string(input)
	fp := fsd.path + "_data"

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
		if fsd.st.MatchFlag(models.USERFLAG_ACCOUNT_UNLOCKED, false) {
			res.FlagReset = append(res.FlagReset, models.USERFLAG_INCORRECTPIN)
			res.FlagSet = append(res.FlagSet, models.USERFLAG_UNLOCKFORUPDATE)
			res.FlagSet = append(res.FlagSet, models.USERFLAG_ACCOUNT_UNLOCKED)
		} else {
			res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
		}
	}
	return res, nil
}

func (fsd *fsData) reset_incorrect_pin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_INCORRECTPIN)
	return res, nil
}

func (fsd *fsData) check_account_status(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}

	//status, err := checkAccountStatus(accountData["TrackingId"])
	status, err := handlers.CheckAccountStatus(accountData["TrackingId"])

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

func (fsd *fsData) quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

func (fsd *fsData) verify_yob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

func (fsd *fsData) reset_incorrect_yob(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, models.USERFLAG_INCORRECTDATEFORMAT)
	return res, nil
}

func (fsd *fsData) check_balance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	fp := fsd.path + "_data"
	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}
	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	balance, err := handlers.CheckBalance(accountData["PublicKey"])
	if err != nil {
		return res, nil
	}
	res.Content = balance

	return res, nil
}

func (fsd *fsData) validate_recipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	recipient := string(input)

	fp := fsd.path + "_data"

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

func (fsd *fsData) transaction_reset(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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

func (fsd *fsData) reset_transaction_amount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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

func (fsd *fsData) max_amount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}

	// mimic a max amount
	res.Content = "10.00"

	return res, nil
}

func (fsd *fsData) validate_amount(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	amount := string(input)

	fp := fsd.path + "_data"

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

func (fsd *fsData) get_recipient(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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

func (fsd *fsData) get_profile_info(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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

func (fsd *fsData) get_sender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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

func (fsd *fsData) quit_with_balance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

	jsonData, err := os.ReadFile(fp)
	if err != nil {
		return res, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return res, err
	}
	balance, err := handlers.CheckBalance(accountData["PublicKey"])
	if err != nil {
		return res, nil
	}
	res.Content = fmt.Sprintf("Your account balance is: %s", balance)
	res.FlagReset = append(res.FlagReset, models.USERFLAG_ACCOUNT_UNLOCKED)
	return res, nil
}

func (fsd *fsData) save_pin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	accountPIN := string(input)
	fp := fsd.path + "_data"

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

func (fsd *fsData) verify_pin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"

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
	st := state.NewState(15)
	st.UseDebug()
	state.FlagDebugger.Register(models.USERFLAG_LANGUAGE_SET, "LANGUAGE_CHANGE")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_CREATED, "ACCOUNT_CREATED")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_SUCCESS, "ACCOUNT_SUCCESS")
	state.FlagDebugger.Register(models.USERFLAG_ACCOUNT_PENDING, "ACCOUNT_PENDING")
	state.FlagDebugger.Register(models.USERFLAG_INCORRECTPIN, "INCORRECTPIN")

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
	fs := &fsData{
		path: fp,
		st:   &st,
	}
	rfs.AddLocalFunc("select_language", fs.set_language)
	rfs.AddLocalFunc("create_account", fs.create_account)
	rfs.AddLocalFunc("save_pin", fs.save_pin)
	rfs.AddLocalFunc("verify_pin", fs.verify_pin)
	rfs.AddLocalFunc("check_identifier", fs.check_identifier)
	rfs.AddLocalFunc("check_account_status", fs.check_account_status)
	rfs.AddLocalFunc("unlock_account", fs.unlock)
	rfs.AddLocalFunc("quit", fs.quit)
	rfs.AddLocalFunc("check_balance", fs.check_balance)
	rfs.AddLocalFunc("validate_recipient", fs.validate_recipient)
	rfs.AddLocalFunc("transaction_reset", fs.transaction_reset)
	rfs.AddLocalFunc("max_amount", fs.max_amount)
	rfs.AddLocalFunc("validate_amount", fs.validate_amount)
	rfs.AddLocalFunc("reset_transaction_amount", fs.reset_transaction_amount)
	rfs.AddLocalFunc("get_recipient", fs.get_recipient)
	rfs.AddLocalFunc("get_sender", fs.get_sender)
	rfs.AddLocalFunc("reset_incorrect", fs.reset_incorrect_pin)
	rfs.AddLocalFunc("save_firstname", fs.save_firstname)
	rfs.AddLocalFunc("save_familyname", fs.save_familyname)
	rfs.AddLocalFunc("save_gender", fs.save_gender)
	rfs.AddLocalFunc("save_location", fs.save_location)
	rfs.AddLocalFunc("save_yob", fs.save_yob)
	rfs.AddLocalFunc("save_offerings", fs.save_offerings)
	rfs.AddLocalFunc("quit_with_balance", fs.quit_with_balance)
	rfs.AddLocalFunc("reset_unlocked", fs.reset_account_unlocked)
	rfs.AddLocalFunc("reset_unlock_for_update", fs.reset_unlock_for_update)
	rfs.AddLocalFunc("get_profile_info", fs.get_profile_info)
	rfs.AddLocalFunc("verify_yob", fs.verify_yob)
	rfs.AddLocalFunc("reset_incorrect_date_format", fs.reset_incorrect_yob)

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

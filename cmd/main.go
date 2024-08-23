package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG_LANGUAGE_SET = iota + state.FLAG_USERSTART
	USERFLAG_ACCOUNT_CREATED
	USERFLAG_ACCOUNT_PENDING
	USERFLAG_ACCOUNT_SUCCESS
	USERFLAG_ACCOUNT_UNLOCKED
	USERFLAG_INVALID_RECIPIENT
	USERFLAG_INVALID_RECIPIENT_WITH_INVITE
	USERFLAG_INCORRECTPIN
	USERFLAG_UNLOCKFORUPDATE
	USERFLAG_INVALID_AMOUNT
	USERFLAG_QUERYPIN
	USERFLAG_VALIDPIN
	USERFLAG_INVALIDPIN
	USERFLAG_UNLOCKFORUPDATE
)

const (
	createAccountURL = "https://custodial.sarafu.africa/api/account/create"
	trackStatusURL   = "https://custodial.sarafu.africa/api/track/"
	checkBalanceURL  = "https://custodial.sarafu.africa/api/account/status/"
)

type accountResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		CustodialId json.Number `json:"custodialId"`
		PublicKey   string      `json:"publicKey"`
		TrackingId  string      `json:"trackingId"`
	} `json:"result"`
}

type trackStatusResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		Transaction struct {
			CreatedAt     time.Time   `json:"createdAt"`
			Status        string      `json:"status"`
			TransferValue json.Number `json:"transferValue"`
			TxHash        string      `json:"txHash"`
			TxType        string      `json:"txType"`
		}
	} `json:"result"`
}

type balanceResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		Balance string      `json:"balance"`
		Nonce   json.Number `json:"nonce"`
	} `json:"result"`
}

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

func (fsd *fsData) saveFirstName(cxt context.Context, sym string, input []byte) (resource.Result, error) {
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

func (fsd *fsData) saveFamilyName(cxt context.Context, sym string, input []byte) (resource.Result, error) {
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
		fmt.Println("FamilyName:", secondname)
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
func (fsd *fsData) saveYOB(cxt context.Context, sym string, input []byte) (resource.Result, error) {
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
		fmt.Println("YOB", yob)
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

func (fsd *fsData) saveLocation(cxt context.Context, sym string, input []byte) (resource.Result, error) {
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
		fmt.Println("Location:", location)
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

func (fsd *fsData) saveGender(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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
		case "1" : gender = "Male"
		case "2" : gender = "Female"
		case "3" : gender = "Other"
		}
		fmt.Println("gender", gender)
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

func (fsd *fsData) saveOfferings(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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
		fmt.Println("Offerings:", offerings)
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

func (fsd *fsData) SetLanguageSelected(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

	res.FlagSet = append(res.FlagSet, USERFLAG_LANGUAGE_SET)

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

	//accountResp, err := createAccount()
	accountResp := accountResponse{
		Ok: true,
		Result: struct {
			CustodialId json.Number `json:"custodialId"`
			PublicKey   string      `json:"publicKey"`
			TrackingId  string      `json:"trackingId"`
		}{
			CustodialId: "636",
			PublicKey:   "0x8d86F9D4A4eae41Dc3B68034895EA97BcA90e8c1",
			TrackingId:  "45c67314-7995-4890-89d6-e5af987754ac",
		}}

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
	res.FlagSet = append(res.FlagSet, USERFLAG_ACCOUNT_CREATED)
	return res, err
}

func (fsd *fsData) resetUnlockForUpdate(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	res.FlagReset = append(res.FlagReset, USERFLAG_UNLOCKFORUPDATE)
	return res, nil
}

func (fsd *fsData) resetAccountUnlocked(ctx context.Context,sym string,input []byte) (resource.Result,error){
  res := resource.Result{}
  st := fsd.st
  isSet := st.MatchFlag(USERFLAG_ACCOUNT_UNLOCKED,true)
  res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
  fmt.Println("ISSET:",isSet)
  return res,nil
}

func (fsd *fsData) checkIdentifier(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

func (fsd *fsData) unLock(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	pin := string(input)
	if len(input) > 0 {
		if pin == "0000" {
			res.FlagSet = append(res.FlagSet, USERFLAG_INCORRECTPIN)
			res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
			return res, nil
		}
		if fsd.st.MatchFlag(USERFLAG_ACCOUNT_UNLOCKED, false) {
			//res.FlagSet = append(res.FlagSet, USERFLAG_UNLOCKFORUPDATE)
			res.FlagSet = append(res.FlagSet, USERFLAG_ACCOUNT_UNLOCKED)
		} else {
			res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
			//res.FlagReset = append(res.FlagReset, USERFLAG_UNLOCKFORUPDATE)
		}
	}
	return res, nil
}

func (fsd *fsData) ResetIncorrectPin(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	isIncorrectPinSet := fsd.st.MatchFlag(USERFLAG_INCORRECTPIN, true)
	if isIncorrectPinSet {
		res.FlagReset = append(res.FlagReset, USERFLAG_INCORRECTPIN)
	} else {
		res.FlagReset = append(res.FlagReset, USERFLAG_INCORRECTPIN)
	}
	return res, nil
}

func (fsd *fsData) ShowUpdateSuccess(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
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

	status, err := checkAccountStatus(accountData["TrackingId"])

	if err != nil {
		fmt.Println("Error checking account status:", err)
		return res, nil
	}

	accountData["Status"] = status

	if status == "REVERTED" {
		res.FlagSet = append(res.FlagSet, USERFLAG_ACCOUNT_SUCCESS)
		res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_PENDING)
	} else {
		res.FlagReset = append(res.FlagSet, USERFLAG_ACCOUNT_SUCCESS)
		res.FlagSet = append(res.FlagReset, USERFLAG_ACCOUNT_PENDING)
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

func createAccount() (*accountResponse, error) {
	resp, err := http.Post(createAccountURL, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accountResp accountResponse
	err = json.Unmarshal(body, &accountResp)
	if err != nil {
		return nil, err
	}

	return &accountResp, nil
}

func checkAccountStatus(trackingId string) (string, error) {
	resp, err := http.Get(trackStatusURL + trackingId)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var trackResp trackStatusResponse
	err = json.Unmarshal(body, &trackResp)
	if err != nil {
		return "", err
	}

	status := trackResp.Result.Transaction.Status

	return status, nil
}

func (fsd *fsData) quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	switch codeFromCtx(ctx) {
	case "swa":
		res.Content = "Asante kwa kutumia huduma ya Sarafu. Kwaheri!"
	default:
		res.Content = "Thank you for using Sarafu. Goodbye!"
	}
	res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
	return res, nil
}

func (fsd *fsData) checkBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

	resp, err := http.Get(checkBalanceURL + accountData["PublicKey"])
	if err != nil {
		return res, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, nil
	}

	var balanceResp balanceResponse
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		return res, nil
	}

	balance := balanceResp.Result.Balance

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
			res.FlagSet = append(res.FlagSet, USERFLAG_INVALID_RECIPIENT)
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

	res.FlagReset = append(res.FlagReset, USERFLAG_INVALID_RECIPIENT, USERFLAG_INVALID_RECIPIENT_WITH_INVITE)

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

	res.FlagReset = append(res.FlagReset, USERFLAG_INVALID_AMOUNT)

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
			res.FlagSet = append(res.FlagSet, USERFLAG_INVALID_AMOUNT)
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


func (fsd *fsData) getProfileInfo(ctx context.Context,sym string,input []byte) (resource.Result,error){
	res :=  resource.Result{}
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
	name :=  accountData["FirstName"]
    gender := accountData["Gender"]
    age := accountData["YOB"]
    location := accountData["Location"]

    // Format the data into a string
    formattedData := fmt.Sprintf("Name: %s\nGender: %s\nAge: %s\nLocation: %s\n", name, gender, age, location)
	

	res.Content = formattedData

	return res,nil
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

func (fsd *fsData) quitWithBalance(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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
	resp, err := http.Get(checkBalanceURL + accountData["PublicKey"])
	if err != nil {
		return res, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, nil
	}

	var balanceResp balanceResponse
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		return res, nil
	}
	balance := balanceResp.Result.Balance
	res.Content = fmt.Sprintf("Your account balance is: %s", balance)
	res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
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
		res.FlagSet = []uint32{USERFLAG_VALIDPIN}
		res.FlagReset = []uint32{USERFLAG_INVALIDPIN}
	} else {
		res.FlagSet = []uint32{USERFLAG_INVALIDPIN}
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
	state.FlagDebugger.Register(USERFLAG_LANGUAGE_SET, "LANGUAGE_CHANGE")
	state.FlagDebugger.Register(USERFLAG_ACCOUNT_CREATED, "ACCOUNT_CREATED")
	state.FlagDebugger.Register(USERFLAG_ACCOUNT_SUCCESS, "ACCOUNT_SUCCESS")
	state.FlagDebugger.Register(USERFLAG_ACCOUNT_PENDING, "ACCOUNT_PENDING")
	state.FlagDebugger.Register(USERFLAG_INCORRECTPIN, "INCORRECTPIN")

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
	rfs.AddLocalFunc("select_language", fs.SetLanguageSelected)
	rfs.AddLocalFunc("create_account", fs.create_account)
	rfs.AddLocalFunc("save_pin", fs.save_pin)
	rfs.AddLocalFunc("verify_pin", fs.verify_pin)
	rfs.AddLocalFunc("check_identifier", fs.checkIdentifier)
	rfs.AddLocalFunc("check_account_status", fs.check_account_status)
	rfs.AddLocalFunc("unlock_account", fs.unLock)
	rfs.AddLocalFunc("quit", fs.quit)
	rfs.AddLocalFunc("check_balance", fs.checkBalance)
	rfs.AddLocalFunc("validate_recipient", fs.validate_recipient)
	rfs.AddLocalFunc("transaction_reset", fs.transaction_reset)
	rfs.AddLocalFunc("max_amount", fs.max_amount)
	rfs.AddLocalFunc("validate_amount", fs.validate_amount)
	rfs.AddLocalFunc("reset_transaction_amount", fs.reset_transaction_amount)
	rfs.AddLocalFunc("get_recipient", fs.get_recipient)
	rfs.AddLocalFunc("get_sender", fs.get_sender)
	rfs.AddLocalFunc("reset_incorrect", fs.ResetIncorrectPin)
	rfs.AddLocalFunc("save_firstname", fs.saveFirstName)
	rfs.AddLocalFunc("save_familyname", fs.saveFamilyName)
	rfs.AddLocalFunc("save_gender", fs.saveGender)
	rfs.AddLocalFunc("save_location", fs.saveLocation)
	rfs.AddLocalFunc("save_yob", fs.saveYOB)
	rfs.AddLocalFunc("save_offerings", fs.saveOfferings)
	rfs.AddLocalFunc("quit_with_balance", fs.quitWithBalance)
	rfs.AddLocalFunc("show_update_success", fs.ShowUpdateSuccess)
	rfs.AddLocalFunc("reset_unlocked",fs.resetAccountUnlocked)
	rfs.AddLocalFunc("reset_unlock_for_update", fs.resetUnlockForUpdate)
	rfs.AddLocalFunc("get_profile_info",fs.getProfileInfo)


	

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

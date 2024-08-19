package main

import (
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
		Balance   string      `json:"balance"`
		Nonce  json.Number `json:"nonce"`
	} `json:"result"`
}

type fsData struct {
	path string
	st *state.State
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

	// accountResp, err := createAccount()

	// if err != nil {
	// 	fmt.Println("Failed to create account:", err)
	// 	return res, err
	// }

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

	res.FlagSet = []uint32{USERFLAG_ACCOUNT_CREATED}
	return res, err
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


func (fsd *fsData) unlock(ctx context.Context,sym string,input []byte) (resource.Result,error){
	res := resource.Result{}
	//res.FlagSet = []uint32{USERFLAG_ACCOUNT_UNLOCKED}
	if fsd.st.MatchFlag(USERFLAG_ACCOUNT_UNLOCKED, false) {
		res.FlagSet = append(res.FlagSet, USERFLAG_ACCOUNT_UNLOCKED)
	} else {
		res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
	}
	return res,nil
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

	if status == "SUCCESS" {
		res.FlagSet = append(res.FlagSet, USERFLAG_ACCOUNT_SUCCESS)
		res.FlagReset = append(res.FlagReset,USERFLAG_ACCOUNT_PENDING)
	} else {
		res.FlagReset = append(res.FlagSet, USERFLAG_ACCOUNT_SUCCESS)
		res.FlagSet = append(res.FlagReset,USERFLAG_ACCOUNT_PENDING)
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

func  (fsd *fsData) quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{
		Content: "",
	}
	st := fsd.st
	if(st.MatchFlag(USERFLAG_ACCOUNT_UNLOCKED,true)){
	     //res.FlagReset = []uint32{USERFLAG_ACCOUNT_UNLOCKED}
		// res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
	}else {
		res.FlagReset = append(res.FlagReset, USERFLAG_ACCOUNT_UNLOCKED)
	}
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
	st := state.NewState(7)
	st.UseDebug()
	state.FlagDebugger.Register(USERFLAG_LANGUAGE_SET, "LANGUAGE_CHANGE")
	state.FlagDebugger.Register(USERFLAG_ACCOUNT_CREATED,"ACCOUNT_CREATED")
	state.FlagDebugger.Register(USERFLAG_ACCOUNT_SUCCESS,"ACCOUNT_SUCCESS")
	state.FlagDebugger.Register(USERFLAG_ACCOUNT_PENDING,"ACCOUNT_PENDING")


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
		st: &st,
	}
	rfs.AddLocalFunc("select_language", fs.SetLanguageSelected)
	rfs.AddLocalFunc("create_account", fs.create_account)
	rfs.AddLocalFunc("check_identifier", fs.checkIdentifier)
	rfs.AddLocalFunc("check_account_status", fs.check_account_status)
	rfs.AddLocalFunc("unlock_account", fs.unlock)
	rfs.AddLocalFunc("quit",fs.quit)
	rfs.AddLocalFunc("check_balance", fs.checkBalance)

	cont, err := en.Init(ctx)
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

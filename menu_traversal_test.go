package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"git.defalsify.org/vise.git/engine"

	"git.grassecon.net/urdt/ussd/enginetest"
)

type TestCase struct {
	Input    []string `json:"input"`
	Expected string   `json:"expected"`
}

type UserRegistration struct {
	UserRegistration []TestCase `json:"user_registration"`
}

type TestData struct {
	UserRegistration []TestCase `json:"user_registration"`
	PinCheck         []TestCase `json:"pincheck"`
}

func TestUserRegistration(t *testing.T) {
	en, pe := enginetest.TestEngine("session1234112")
	w := bytes.NewBuffer(nil)
	file, err := os.Open("test_data.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var testData TestData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&testData); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	var inputBuilder strings.Builder
	for _, testCase := range testData.UserRegistration {
		inputBuilder.WriteString(strings.Join(testCase.Input, "\n") + "\n")
	}
	readers := bufio.NewReader(strings.NewReader(inputBuilder.String()))
	engine.Loop(context.Background(), en, readers, w, nil)
	st := pe.GetState()
	sym, _ := st.Where()
	b := w.Bytes()
    fmt.Println("Menu:",string(b))
	fmt.Println("Rendering symbol:", sym)

}

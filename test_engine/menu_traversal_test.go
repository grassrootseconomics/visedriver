package main

import (
	"bytes"
	"context"
	"log"
	"math/rand"
	"os"
	"regexp"
	"testing"
	"time"

	"git.grassecon.net/urdt/ussd/driver"
	enginetest "git.grassecon.net/urdt/ussd/engine"
	"github.com/gofrs/uuid"
)

var (
	testData  = driver.ReadData()
	testStore = ".test_state"
	sessionID string
	src       = rand.NewSource(42)
	g         = rand.New(src)
)

func GenerateSessionId() string {
	uu := uuid.NewGenWithOptions(uuid.WithRandomReader(g))
	v, err := uu.NewV4()
	if err != nil {
		panic(err)
	}
	return v.String()
}

// Extract the public key from the engine response
func extractPublicKey(response []byte) string {
	// Regex pattern to match the public key starting with 0x and 40 characters
	re := regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
	match := re.Find(response)
	if match != nil {
		return string(match)
	}
	return ""
}

func TestMain(m *testing.M) {
	sessionID = GenerateSessionId()
	testStore = ".test_state"
	defer func() {
		if err := os.RemoveAll(testStore); err != nil {
			log.Fatalf("Failed to delete state store %s: %v", testStore, err)
		}
	}()
	m.Run()
}

func TestAccountCreationSuccessful(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_successful")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fatalf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				_, err = en.Flush(ctx, w)
				if err != nil {
					t.Fatalf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
	// Adding a sleep after the test to wait for registration to complete the process
	time.Sleep(5 * time.Second)
}

func TestAccountRegistrationRejectTerms(t *testing.T) {
	// Generate a new UUID for this edge case test
	uu := uuid.NewGenWithOptions(uuid.WithRandomReader(g))
	v, err := uu.NewV4()
	if err != nil {
		t.Fail()
	}
	edgeCaseSessionID := v.String()
	en, fn := enginetest.TestEngine(edgeCaseSessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_reject_terms")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fatalf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Fatalf("Test case '%s' failed during Flush: %v", group.Name, err)
				}

				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestSendWithInvalidInputs(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "send_with_invalid_inputs")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fatalf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Fatalf("Test case '%s' failed during Flush: %v", group.Name, err)
				}

				b := w.Bytes()

				// Extract the dynamic public key from the output
				publicKey := extractPublicKey(b)

				// Replace placeholder {public_key} with the actual dynamic public key
				expectedContent := bytes.Replace([]byte(step.ExpectedContent), []byte("{public_key}"), []byte(publicKey), -1)

				if !bytes.Equal(b, expectedContent) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Check_My_Balance(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_check_my_balance")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMainMenuHelp(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "main_menu_help")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fatalf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Fatalf("Test case '%s' failed during Flush: %v", group.Name, err)
				}

				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMainMenuQuit(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "main_menu_quit")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fatalf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Fatalf("Test case '%s' failed during Flush: %v", group.Name, err)
				}

				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Check_Community_Balance(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_check_community_balance")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccountChangePin(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "my_account_change_pin")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fatalf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Fatalf("Test case '%s' failed during Flush: %v", group.Name, err)
				}

				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Change_Language(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_language_change")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Edit_firstname(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_edit_firstname")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Edit_familyname(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_edit_familyname")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Edit_gender(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_edit_gender")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Edit_yob(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_edit_yob")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Edit_location(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_edit_location")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_Edit_offerings(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_edit_offerings")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)

				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_MyAddress(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_my_address")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()

				publicKey := extractPublicKey(b)

				expectedContent := bytes.Replace([]byte(step.ExpectedContent), []byte("{public_key}"), []byte(publicKey), -1)

				if !bytes.Equal(b, expectedContent) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_View_Profile(t *testing.T) {
	en, fn := enginetest.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_view_profile")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				cont, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
				}
				if !cont {
					break
				}
				w := bytes.NewBuffer(nil)
				if _, err := en.Flush(ctx, w); err != nil {
					t.Errorf("Test case '%s' failed during Flush: %v", group.Name, err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

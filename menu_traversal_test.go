package main

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"git.grassecon.net/urdt/ussd/driver"
	enginetest "git.grassecon.net/urdt/ussd/engine"
)

var (
	testData = driver.ReadData()
)

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

func TestUserRegistration(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112")
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
}

func TestTerms(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112_a")
	defer fn()
	ctx := context.Background()
	sessions := testData

	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_accept_terms")
		for _, group := range groups {
			for _, step := range group.Steps {
				_, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fatalf("Test case '%s' failed during Exec: %v", group.Name, err)
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
}

func TestAccountRegistrationRejectTerms(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112_b")
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

func TestAccountRegistrationInvalidPin(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112_c")
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_invalid_pin")
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
	en, fn := enginetest.TestEngine("session1234112")
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

func TestMainMenuHelp(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112")
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
	en, fn := enginetest.TestEngine("session1234112")
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

func TestMyAccountChangePin(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112")
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
	en, fn := enginetest.TestEngine("session1234112")
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

func TestMyAccount_Savefirstname(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112")
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "menu_my_account_save_firstname")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
				_, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Errorf("Test case '%s' failed at input '%s': %v", group.Name, step.Input, err)
					return
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

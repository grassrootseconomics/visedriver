package main

import (
	"bytes"
	"context"
	"testing"

	"git.grassecon.net/urdt/ussd/driver"
	enginetest "git.grassecon.net/urdt/ussd/engine"
)

var (
	testData = driver.ReadData()
)

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
	en, fn := enginetest.TestEngine("session1234112")
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

func TestSendWithInvalidRecipient(t *testing.T) {
	en, fn := enginetest.TestEngine("session1234112")
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "send_with_invalid_recipient")
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

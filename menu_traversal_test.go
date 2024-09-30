package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"git.grassecon.net/urdt/ussd/driver"
	enginetest "git.grassecon.net/urdt/ussd/engine"
)

var (
	testData = driver.ReadData()
)

func TestUserRegistration(t *testing.T) {
	en, _ := enginetest.TestEngine("session1234112")
	defer en.Finish()
	//var err error
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_successful")
		for _, group := range groups {
			for _, step := range group.Steps {
				//			for {
				_, err := en.Exec(ctx, []byte(step.Input))

				if err != nil {
					t.Fail()
				}
				w := bytes.NewBuffer(nil)
				_, err = en.Flush(ctx, w)
				if err != nil {
					t.Fatal(err)
				}
				b := w.Bytes()
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
				//				}

			}
		}
	}
}

func TestTerms(t *testing.T) {
	en, _ := enginetest.TestEngine("session1234112")
	defer en.Finish()
	ctx := context.Background()
	sessions := testData

	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_accept_terms")
		for _, group := range groups {
			for _, step := range group.Steps {
				//			for {
				_, err := en.Exec(ctx, []byte(step.Input))
				if err != nil {
					t.Fail()
				}

				w := bytes.NewBuffer(nil)
				_, err = en.Flush(ctx, w)
				if err != nil {
					t.Fatal(err)
				}
				b := w.Bytes()
				fmt.Println("valuehere:", string(b))
				if !bytes.Equal(b, []byte(step.ExpectedContent)) {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
				//			}

			}
		}
	}
}

func TestAccountRegistrationRejectTerms(t *testing.T) {
	en, _ := enginetest.TestEngine("session1234112")
	var err error
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_reject_terms")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, _ := en.Exec(ctx, []byte(step.Input))
				if cont {
					w := bytes.NewBuffer(nil)
					_, err = en.Flush(ctx, w)
					if err != nil {
						t.Fatal(err)
					}
					b := w.Bytes()
					if !bytes.Equal(b, []byte(step.ExpectedContent)) {
						t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
					}

				}
			}
		}
	}
}

func TestAccountRegistrationInvalidPin(t *testing.T) {
	en, _ := enginetest.TestEngine("session1234112")
	var err error
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_invalid_pin")
		for _, group := range groups {
			for _, step := range group.Steps {
				cont, _ := en.Exec(ctx, []byte(step.Input))
				if cont {
					w := bytes.NewBuffer(nil)
					_, err = en.Flush(ctx, w)
					if err != nil {
						t.Fatal(err)
					}
					b := w.Bytes()
					if !bytes.Equal(b, []byte(step.ExpectedContent)) {
						t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
					}

				}
			}
		}
	}
}

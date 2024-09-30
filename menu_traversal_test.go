package main

import (
	"bytes"
	"context"
	"testing"

	"git.grassecon.net/urdt/ussd/driver"
	"git.grassecon.net/urdt/ussd/engine"
)

var (
	testData = driver.ReadData()
)

func TestUserRegistration(t *testing.T) {
	en, _ := enginetest.TestEngine("session1234112")
	var err error
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_successful")
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

func TestAcceptTerms(t *testing.T) {
	en, _ := enginetest.TestEngine("session12341123")
	var err error
	ctx := context.Background()
	sessions := testData

	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "account_creation_accept_terms")
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

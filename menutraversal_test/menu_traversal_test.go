package menutraversaltest

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"testing"

	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/internal/testutil"
	"git.grassecon.net/urdt/ussd/internal/testutil/driver"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testData  = driver.ReadData()
	testStore = ".test_state"
	sessionID string
	src       = rand.NewSource(42)
	g         = rand.New(src)
)

var groupTestFile = flag.String("test-file", "group_test.json", "The test file to use for running the group tests")
var database = flag.String("db", "gdbm", "Specify the database (gdbm or postgres)")
var dbSchema = flag.String("schema", "test", "Specify the database schema (default test)")

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

// Extracts the balance value from the engine response.
func extractBalance(response []byte) string {
	// Regex to match "Balance: <amount> <symbol>" followed by a newline
	re := regexp.MustCompile(`(?m)^Balance:\s+(\d+(\.\d+)?)\s+([A-Z]+)`)
	match := re.FindSubmatch(response)
	if match != nil {
		return string(match[1]) + " " + string(match[3]) // "<amount> <symbol>"
	}
	return ""
}

// Extracts the Maximum amount value from the engine response.
func extractMaxAmount(response []byte) string {
	// Regex to match "Maximum amount: <amount>" followed by a newline
	re := regexp.MustCompile(`(?m)^Maximum amount:\s+(\d+(\.\d+)?)`)
	match := re.FindSubmatch(response)
	if match != nil {
		return string(match[1]) // "<amount>"
	}
	return ""
}

// Extracts the send amount value from the engine response.
func extractSendAmount(response []byte) string {
	// Regex to match the pattern "will receive X.XX SYM from"
	re := regexp.MustCompile(`will receive (\d+\.\d{2}\s+[A-Z]+) from`)
	match := re.FindSubmatch(response)
	if match != nil {
		return string(match[1]) // Returns "X.XX SYM"
	}
	return ""
}

func TestMain(m *testing.M) {
	// Parse the flags
	flag.Parse()

	sessionID = GenerateSessionId()
	defer func() {
		if err := os.RemoveAll(testStore); err != nil {
			log.Fatalf("Failed to delete state store %s: %v", testStore, err)
		}
	}()

	// Set the selected database
	testutil.SetDatabase(*database, *dbSchema)

	// Cleanup the schema table after tests
	defer func() {
		if *database == "postgres" {
			ctx := context.Background()
			connStr := storage.BuildConnStr()
			dbConn, err := pgxpool.New(ctx, connStr)
			if err != nil {
				log.Fatalf("Failed to connect to database for cleanup: %v", err)
			}
			defer dbConn.Close()

			query := fmt.Sprintf("DELETE FROM %s.kv_vise;", *dbSchema)
			_, execErr := dbConn.Exec(ctx, query)
			if execErr != nil {
				log.Printf("Failed to cleanup table %s.kv_vise: %v", *dbSchema, execErr)
			} else {
				log.Printf("Successfully cleaned up table %s.kv_vise", *dbSchema)
			}
		}
	}()

	m.Run()
}

func TestAccountCreationSuccessful(t *testing.T) {
	en, fn, eventChannel := testutil.TestEngine(sessionID)
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
				match, err := step.MatchesExpectedContent(b)
				if err != nil {
					t.Fatalf("Error compiling regex for step '%s': %v", step.Input, err)
				}
				if !match {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
	<-eventChannel
}

func TestAccountRegistrationRejectTerms(t *testing.T) {
	// Generate a new UUID for this edge case test
	uu := uuid.NewGenWithOptions(uuid.WithRandomReader(g))
	v, err := uu.NewV4()
	if err != nil {
		t.Fail()
	}
	edgeCaseSessionID := v.String()
	en, fn, _ := testutil.TestEngine(edgeCaseSessionID)
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
				match, err := step.MatchesExpectedContent(b)
				if err != nil {
					t.Fatalf("Error compiling regex for step '%s': %v", step.Input, err)
				}
				if !match {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMainMenuHelp(t *testing.T) {
	en, fn, _ := testutil.TestEngine(sessionID)
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
				balance := extractBalance(b)

				expectedContent := []byte(step.ExpectedContent)
				expectedContent = bytes.Replace(expectedContent, []byte("{balance}"), []byte(balance), -1)

				step.ExpectedContent = string(expectedContent)
				match, err := step.MatchesExpectedContent(b)
				if err != nil {
					t.Fatalf("Error compiling regex for step '%s': %v", step.Input, err)
				}
				if !match {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMainMenuQuit(t *testing.T) {
	en, fn, _ := testutil.TestEngine(sessionID)
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
				balance := extractBalance(b)

				expectedContent := []byte(step.ExpectedContent)
				expectedContent = bytes.Replace(expectedContent, []byte("{balance}"), []byte(balance), -1)

				step.ExpectedContent = string(expectedContent)
				match, err := step.MatchesExpectedContent(b)
				if err != nil {
					t.Fatalf("Error compiling regex for step '%s': %v", step.Input, err)
				}
				if !match {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestMyAccount_MyAddress(t *testing.T) {
	en, fn, _ := testutil.TestEngine(sessionID)
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

				balance := extractBalance(b)
				publicKey := extractPublicKey(b)

				expectedContent := []byte(step.ExpectedContent)
				expectedContent = bytes.Replace(expectedContent, []byte("{balance}"), []byte(balance), -1)
				expectedContent = bytes.Replace(expectedContent, []byte("{public_key}"), []byte(publicKey), -1)

				step.ExpectedContent = string(expectedContent)
				match, err := step.MatchesExpectedContent(b)
				if err != nil {
					t.Fatalf("Error compiling regex for step '%s': %v", step.Input, err)
				}
				if !match {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expectedContent, b)
				}
			}
		}
	}
}

func TestMainMenuSend(t *testing.T) {
	en, fn, _ := testutil.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	sessions := testData
	for _, session := range sessions {
		groups := driver.FilterGroupsByName(session.Groups, "send_with_invite")
		for _, group := range groups {
			for index, step := range group.Steps {
				t.Logf("step %v with input %v", index, step.Input)
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
				balance := extractBalance(b)
				max_amount := extractMaxAmount(b)
				send_amount := extractSendAmount(b)

				expectedContent := []byte(step.ExpectedContent)
				expectedContent = bytes.Replace(expectedContent, []byte("{balance}"), []byte(balance), -1)
				expectedContent = bytes.Replace(expectedContent, []byte("{max_amount}"), []byte(max_amount), -1)
				expectedContent = bytes.Replace(expectedContent, []byte("{send_amount}"), []byte(send_amount), -1)
				expectedContent = bytes.Replace(expectedContent, []byte("{session_id}"), []byte(sessionID), -1)

				step.ExpectedContent = string(expectedContent)
				match, err := step.MatchesExpectedContent(b)
				if err != nil {
					t.Fatalf("Error compiling regex for step '%s': %v", step.Input, err)
				}
				if !match {
					t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", step.ExpectedContent, b)
				}
			}
		}
	}
}

func TestGroups(t *testing.T) {
	groups, err := driver.LoadTestGroups(*groupTestFile)
	if err != nil {
		log.Fatalf("Failed to load test groups: %v", err)
	}
	en, fn, _ := testutil.TestEngine(sessionID)
	defer fn()
	ctx := context.Background()
	// Create test cases from loaded groups
	tests := driver.CreateTestCases(groups)
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cont, err := en.Exec(ctx, []byte(tt.Input))
			if err != nil {
				t.Errorf("Test case '%s' failed at input '%s': %v", tt.Name, tt.Input, err)
				return
			}
			if !cont {
				return
			}
			w := bytes.NewBuffer(nil)
			if _, err := en.Flush(ctx, w); err != nil {
				t.Errorf("Test case '%s' failed during Flush: %v", tt.Name, err)
			}
			b := w.Bytes()
			balance := extractBalance(b)

			expectedContent := []byte(tt.ExpectedContent)
			expectedContent = bytes.Replace(expectedContent, []byte("{balance}"), []byte(balance), -1)

			tt.ExpectedContent = string(expectedContent)

			match, err := tt.MatchesExpectedContent(b)
			if err != nil {
				t.Fatalf("Error compiling regex for step '%s': %v", tt.Input, err)
			}
			if !match {
				t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", tt.ExpectedContent, b)
			}
		})
	}
}

package driver

import (
	"encoding/json"
	"os"
)

type StepTest struct {
	Input           string `json:"input"`
	ExpectedContent string `json:"expectedContent"`
}

// Group represents a group of steps.
type GroupTest struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

// DataGroup represents the overall structure of the JSON.
type DataGroup struct {
	Groups []Group `json:"groups"`
}

func LoadTestGroups(filePath string) (DataGroup, error) {
	var sessionsData DataGroup
	data, err := os.ReadFile(filePath)
	if err != nil {
		return sessionsData, err
	}
	err = json.Unmarshal(data, &sessionsData)
	return sessionsData, err
}

func CreateTestCases(group DataGroup) []struct {
	Name            string
	Input           string
	ExpectedContent string
} {
	var tests []struct {
		Name            string
		Input           string
		ExpectedContent string
	}
	for _, group := range group.Groups {
		for _, step := range group.Steps {
			// Create a test case for each group
			tests = append(tests, struct {
				Name            string
				Input           string
				ExpectedContent string
			}{
				Name:            group.Name,
				Input:           step.Input,
				ExpectedContent: step.ExpectedContent,
			})
		}
	}

	return tests
}

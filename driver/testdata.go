package driver

import (
	"encoding/json"
	"log"
	"os"
)

type Step struct {
	Input           string `json:"input"`
	ExpectedContent string `json:"expectedContent"`
}

type Group struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

type Session struct {
	Name   string  `json:"name"`
	Groups []Group `json:"groups"`
}

func ReadData() []Session {
	data, err := os.ReadFile("test_data.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	// Unmarshal JSON data
	var sessions []Session
	err = json.Unmarshal(data, &sessions)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	return sessions
}

func FilterGroupsByName(groups []Group, name string) []Group {
	var filteredGroups []Group
	for _, group := range groups {
		if group.Name == name {
			filteredGroups = append(filteredGroups, group)
		}
	}
	return filteredGroups
}

func Map[T any, U any](input []T, fn func(T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = fn(v)
	}
	return result
}

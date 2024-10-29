package utils

import (
	"bufio"
	"os"
	"strconv"
)

type AdminStore struct {
	filePath string
}

// Creates a new Admin store
func NewAdminStore(filePath string) *AdminStore {
	return &AdminStore{filePath: filePath}
}

// Seed saves a list of phonumbers with admin privileges
func (as *AdminStore) Seed() error {
	adminNumbers := []int64{254705136690, 123456789012, 987654321098}
	file, err := os.Create(as.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, num := range adminNumbers {
		_, err := writer.WriteString(strconv.FormatInt(num, 10) + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}


func (as *AdminStore) load() ([]int64, error) {
	file, err := os.Open(as.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var numbers []int64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		num, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			return nil, err
		}
		numbers = append(numbers, num)
	}
	return numbers, scanner.Err()
}

func (as *AdminStore) IsAdmin(phoneNumber int64) (bool, error) {
	phoneNumbers, err := as.load()
	if err != nil {
		return false, err
	}
	for _, phonenumber := range phoneNumbers {
		if phonenumber == phoneNumber {
			return true, nil
		}
	}
	return false, nil
}

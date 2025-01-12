package env

import (
	"log"
	"os"
	"path"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	LoadEnvVariablesPath(".")
}

func LoadEnvVariablesPath(dir string) {
	fp := path.Join(dir, ".env")
	err := godotenv.Load(fp)
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

// Helper to get environment variables with a default fallback
func GetEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// Helper to safely convert environment variables to uint
func GetEnvUint(key string, defaultVal uint) uint {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			return uint(parsed)
		}
	}
	return defaultVal
}

package config

import (
	"strings"

	"git.grassecon.net/grassrootseconomics/visedriver/env"
)

var (
	defaultLanguage		   = "eng"
	languages []string
)

var (
	DbConn		string
	ResourceDbConn	string
	UserDbConn	string
	StateDbConn	string
	DefaultLanguage	    string
	Languages	[]string
)

func setLanguage() error {
	defaultLanguage = env.GetEnv("DEFAULT_LANGUAGE", defaultLanguage)
	languages = strings.Split(env.GetEnv("LANGUAGES", defaultLanguage), ",")
	haveDefaultLanguage := false
	for i, v := range(languages) {
		languages[i] = strings.ReplaceAll(v, " ", "")
		if languages[i] == defaultLanguage {
			haveDefaultLanguage = true
		}
	}

	if !haveDefaultLanguage {
		languages = append([]string{defaultLanguage}, languages...)
	}

	return nil
}



func setConn() error {
	DbConn = env.GetEnv("DB_CONN", "")
	UserDbConn = env.GetEnv("DB_CONN_USER", DbConn)
	StateDbConn = env.GetEnv("DB_CONN_STATE", DbConn)
	ResourceDbConn = env.GetEnv("DB_CONN_RESOURCE", DbConn)
	return nil
}

// LoadConfig initializes the configuration values after environment variables are loaded.
func LoadConfig() error {
	err := setConn()
	if err != nil {
		return err
	}
	err = setLanguage()
	if err != nil {
		return err
	}
	DefaultLanguage = defaultLanguage
	Languages = languages

	return nil
}

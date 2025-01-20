package config

import (
	"strings"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/env"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
)

var (
	logg = logging.NewVanilla().WithDomain("visedriver-config")
	defaultLanguage		   = "eng"
	languages []string
	DefaultLanguage	    string
	dbConn	string
	dbConnMissing	bool
	stateDbConn	string
	resourceDbConn	string
	userDbConn	string
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
	dbConn = env.GetEnv("DB_CONN", "?")
	stateDbConn = env.GetEnv("DB_CONN_STATE", dbConn)
	resourceDbConn = env.GetEnv("DB_CONN_RESOURCE", dbConn)
	userDbConn = env.GetEnv("DB_CONN_USER", dbConn)
	return nil
}

func ApplyConn(connStr *string, stateConnStr *string, resourceConnStr *string, userConnStr *string) {
	if connStr != nil {
		dbConn = *connStr
	}
	if stateConnStr != nil {
		stateDbConn = *stateConnStr
	}
	if resourceConnStr != nil {
		resourceDbConn = *resourceConnStr
	}
	if userConnStr != nil {
		userDbConn = *userConnStr
	}

	if dbConn == "?" {
		dbConn = ""
	}

	if stateDbConn == "?" {
		stateDbConn = dbConn
	}
	if resourceDbConn == "?" {
		resourceDbConn = dbConn
	}
	if userDbConn == "?" {
		userDbConn = dbConn
	}
}

func GetConns() (storage.Conns, error) {
	o := storage.NewConns()
	c, err := storage.ToConnData(stateDbConn)
	if err != nil {
		return o, err
	}
	o.Set(c, storage.STORETYPE_STATE)
	c, err = storage.ToConnData(resourceDbConn)
	if err != nil {
		return o, err
	}
	o.Set(c, storage.STORETYPE_RESOURCE)
	c, err = storage.ToConnData(userDbConn)
	if err != nil {
		return o, err
	}
	o.Set(c, storage.STORETYPE_USER)
	return o, nil
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

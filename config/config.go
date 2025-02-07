package config

import (
	"strconv"
	"strings"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/env"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
)

var (
	logg               = logging.NewVanilla().WithDomain("visedriver-config")
	defaultLanguage    = "eng"
	languages          []string
	DefaultLanguage    string
	dbConn             string
	dbConnMissing      bool
	dbConnMode         storage.DbMode
	stateDbConn        string
	stateDbConnMode    storage.DbMode
	resourceDbConn     string
	resourceDbConnMode storage.DbMode
	userDbConn         string
	userDbConnMode     storage.DbMode
	Languages          []string
	configManager      *Config
)

type Override struct {
	DbConn           string
	DbConnMode       storage.DbMode
	StateConn        string
	StateConnMode    storage.DbMode
	ResourceConn     string
	ResourceConnMode storage.DbMode
	UserConn         string
	UserConnMode     storage.DbMode
}

func setLanguage() error {
	defaultLanguage = env.GetEnv("DEFAULT_LANGUAGE", defaultLanguage)
	languages = strings.Split(env.GetEnv("LANGUAGES", defaultLanguage), ",")
	haveDefaultLanguage := false
	for i, v := range languages {
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

func ApplyConn(override *Override) {
	if override.DbConn != "?" {
		dbConn = override.DbConn
		stateDbConn = override.StateConn
		resourceDbConn = override.ResourceConn
		userDbConn = override.UserConn
	}
	dbConnMode = override.DbConnMode
	if override.StateConn != "?" {
		stateDbConn = override.StateConn
	}
	if override.ResourceConn != "?" {
		resourceDbConn = override.ResourceConn
	}
	if override.UserConn != "?" {
		userDbConn = override.UserConn
	}

	if dbConn == "?" {
		dbConn = ""
	}

	if stateDbConn == "?" {
		stateDbConn = dbConn
		stateDbConnMode = dbConnMode
	}
	if resourceDbConn == "?" {
		resourceDbConn = dbConn
		resourceDbConnMode = dbConnMode
	}
	if userDbConn == "?" {
		userDbConn = dbConn
		userDbConnMode = dbConnMode
	}

	logg.Debugf("conns", "conn", dbConn, "user", userDbConn)
	if override.DbConnMode != storage.DBMODE_ANY {
		dbConnMode = override.DbConnMode
	}
	if override.StateConnMode != storage.DBMODE_ANY {
		stateDbConnMode = override.StateConnMode
	}
	if override.ResourceConnMode != storage.DBMODE_ANY {
		resourceDbConnMode = override.ResourceConnMode
	}
	if override.UserConnMode != storage.DBMODE_ANY {
		userDbConnMode = override.UserConnMode
	}
}

func GetConns() (storage.Conns, error) {
	o := storage.NewConns()
	c, err := storage.ToConnDataMode(stateDbConn, stateDbConnMode)
	if err != nil {
		return o, err
	}
	o.Set(c, storage.STORETYPE_STATE)
	c, err = storage.ToConnDataMode(resourceDbConn, resourceDbConnMode)
	if err != nil {
		return o, err
	}
	o.Set(c, storage.STORETYPE_RESOURCE)
	c, err = storage.ToConnDataMode(userDbConn, userDbConnMode)
	if err != nil {
		return o, err
	}
	o.Set(c, storage.STORETYPE_USER)
	return o, nil
}

// LoadConfig initializes the configuration values after environment variables are loaded.
func LoadConfig() error {
	configManager = NewConfig(logg)

	// Add configuration keys with validation
	configManager.AddKey("HOST", "127.0.0.1", false, nil)
	configManager.AddKey("PORT", "7123", false, func(v string) error {
		_, err := strconv.Atoi(v)
		return err
	})
	configManager.AddKey("DB_CONN", "", true, nil)
	// ... add other keys ?  or is enough :/ ...

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

	// Report configuration
	configManager.Report("INFO")
	return nil
}

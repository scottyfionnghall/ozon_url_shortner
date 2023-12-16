package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/scottyfionnghall/ozonurlshortener/internal/storage"
	"go.uber.org/zap"
)

var (
	InMemoryDB = false
)

type SQLDBVars struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     string
	SSLMode  string
}

func NewLogger() (*zap.Logger, error) {
	rawJSON := []byte(`{
		"level": "debug",
		"encoding": "json",
		"outputPaths": ["stdout", "./logs"],
		"errorOutputPaths": ["stderr","./errors"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	return cfg.Build()
}

func main() {
	logger, err := NewLogger()
	if err != nil {
		return
	}

	sqldbvars := SQLDBVars{
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_SSLMODE"),
	}

	databaseChoice := flag.Bool("m", false, "if added, will use in memmory database and not postgres")
	flag.Parse()
	InMemoryDB = *databaseChoice

	addr := os.Getenv("API_ADDR")
	store, err := GetDB(addr, logger, sqldbvars)
	if err != nil {
		logger.Error("failed to initialize datbase", zap.String("error", err.Error()))
	}

	server := NewAPISever(addr, store, *logger)
	err = server.Run()
	if err != nil {
		logger.Error(err.Error())
	}
}

func GetDB(addr string, logger *zap.Logger, sqldbvars SQLDBVars) (storage.Storage, error) {
	switch InMemoryDB {
	case true:
		store := storage.NewCache()
		err := store.Init()
		if err != nil {
			logger.Error(err.Error())
		}

		logger.Info("API Server with in-memory db running!", zap.String("addr", addr))
		return store, nil
	case false:

		connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
			sqldbvars.User, sqldbvars.Password, sqldbvars.DBName, sqldbvars.Host, sqldbvars.Port, sqldbvars.SSLMode)

		store, err := storage.NewPostgresStore(connStr)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		// defer store.DB.Close()

		if err := store.Init(); err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		logger.Info("API Server with postgres db running!", zap.String("addr", addr))
		return store, nil
	}

	return nil, fmt.Errorf("could not parse flag")
}

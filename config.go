package main

type Database struct {
	DatabaseName string
	UserName     string
	Server       string
	Port         string
	SslMode      string
}

type AppConfig struct {
	Database Database
}

func NewAppConfig() (*AppConfig, error) {
	db := Database{
		DatabaseName: "pgtest",
		UserName:     "pgtest",
		Server:       "localhost",
		Port:         "5432",
		SslMode:      "disable",
	}
	config := AppConfig{
		Database: db,
	}
	return &config, nil
}

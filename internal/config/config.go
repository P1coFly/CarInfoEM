package config

import (
	"os"
)

type config struct {
	Env         string
	HostDb      string
	PortDB      string
	UserDB      string
	PasswordDB  string
	NameDb      string
	HostCarInfo string
	Server
}

type Server struct {
	Port string
}

func MustLoad() *config {

	return &config{Env: os.Getenv("ENV"), HostDb: os.Getenv("HOST_DB"), PortDB: os.Getenv("PORT_DB"),
		UserDB: os.Getenv("USER_DB"), PasswordDB: os.Getenv("PASSWORD_DB")}
}

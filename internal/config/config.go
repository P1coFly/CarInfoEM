package config

import (
	"os"
)

type config struct {
	Env            string
	HostDB         string
	PortDB         string
	UserDB         string
	PasswordDB     string
	NameDB         string
	HostCarInfo    string
	MigrationsPath string
	Server
}

type Server struct {
	Port string
}

func MustLoad() *config {

	return &config{Env: os.Getenv("ENV"), HostDB: os.Getenv("HOST_DB"), PortDB: os.Getenv("PORT_DB"),
		UserDB: os.Getenv("USER_DB"), PasswordDB: os.Getenv("PASSWORD_DB"), NameDB: os.Getenv("NAME_DB"),
		HostCarInfo: os.Getenv("HOST_CARINFO"), MigrationsPath: os.Getenv("MIGRATIONS_PATH"),
		Server: Server{Port: os.Getenv("PORT")}}
}

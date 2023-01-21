package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type Config struct {
	Server struct {
		Address string
		Port    string
		Timeout int
		Prod    bool
	}
	Redis struct {
		Address string
		Port    string
		DB      int
		Timeout int
	}
	Mpesa struct {
		ConsumerKey    string
		ConsumerSecret string
		PassKey        string
		BusinessName   string
		BusinessDesc   string
		ShortCode      string
		CallbackURL    string
		Timeout        int
	}
}

func MustLoad(loc string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(loc, &config); err != nil {
		return config, errors.Wrap(err, "Unable to decode config")
	}

	return config, nil
}

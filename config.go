package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ListenAddress string   `toml:"listen"`
	StashHost     string   `toml:"stash_host"`
	StashUser     string   `toml:"stash_user"`
	StashPassword string   `toml:"stash_password"`
	LintArgs      []string `toml:"lint_args"`
}

func getConfig(path string) (Config, error) {
	config := Config{}

	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return config, err
	}

	validationEmptyError := "`%s` value can't be empty"
	switch "" {
	case config.ListenAddress:
		return config, fmt.Errorf(validationEmptyError, "listen")

	case config.StashHost:
		return config, fmt.Errorf(validationEmptyError, "stash_host")

	case config.StashUser:
		return config, fmt.Errorf(validationEmptyError, "stash_user")

	case config.StashPassword:
		return config, fmt.Errorf(validationEmptyError, "stash_password")
	}

	return config, nil
}

package config

import (
	"flag"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/spf13/viper"
)

var (
	configPath = flag.String("config", "", "path for viper config file")
	configType = flag.String("config_type", "yaml", "config type that will be used to parse config file")
)

func NewConfig() (Config, error) {
	if *configPath == "" {
		return nil, errors.New("config path not defined")
	}

	config := viper.New()

	config.SetConfigFile(*configPath)
	config.SetConfigType(*configType)

	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	return adapter{Viper: config}, nil
}

type adapter struct {
	*viper.Viper
}

func (a adapter) UnmarshalKey(key string, rawVal interface{}) error {
	return a.Viper.UnmarshalKey(key, rawVal)
}

func (a adapter) Unmarshal(rawVal interface{}) error {
	return a.Viper.Unmarshal(rawVal)
}

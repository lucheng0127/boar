package config

import "github.com/spf13/viper"

type ApiConfigSet struct {
	Port int `mapstructure:"port"`
}

type BoarConfigSet struct {
	Api ApiConfigSet `mapstructure:"api"`
}

func ReadConfigFile(path, format string) (*BoarConfigSet, error) {
	config := new(BoarConfigSet)
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType(format)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := v.UnmarshalExact(config); err != nil {
		return nil, err
	}
	return config, nil
}

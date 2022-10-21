package config

import "github.com/spf13/viper"

type ApiConfigSet struct {
	Port int `mapstructure:"port"`
}

type AgentConfigSet struct {
	Host string `mapstructure:"host"`
}

type BoarConfigSet struct {
	Api   ApiConfigSet   `mapstructure:"api"`
	Agent AgentConfigSet `mapstructure:"agent"`
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

package config

import "github.com/spf13/viper"

type GRPC struct {
	Host string
	Port int
}
type SubPub struct{ Buffer int }
type Log struct{ Level string }

type Config struct {
	GRPC   GRPC
	SubPub SubPub
	Log    Log
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var c Config
	return &c, viper.Unmarshal(&c)
}

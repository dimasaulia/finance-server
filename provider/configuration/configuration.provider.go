package configuration

import "github.com/spf13/viper"

type IConfigutation interface {
	LoadEnv() *viper.Viper
}

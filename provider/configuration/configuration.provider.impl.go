package configuration

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
)

type ConfigurationImpl struct {
	FileName string
}

func NewConfiguration(fileName string) IConfigutation {
	return &ConfigurationImpl{
		FileName: fileName,
	}
}

func (c *ConfigurationImpl) LoadEnv() *viper.Viper {
	config := viper.New()
	config.SetConfigFile(".env")
	config.AddConfigPath(".")

	err := config.ReadInConfig()
	if err != nil {
		log.Error("Failed to read configuration file \n")
	}

	return config
}

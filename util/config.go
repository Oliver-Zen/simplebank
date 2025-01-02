package util

import (
	"github.com/spf13/viper"
)

// `config` stores all configurations of the application.
// The values are read by Viper from a config file or environment variables.
type config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoakConfig reads configuration from the file or environment variables.
func LoadConfig(path string) (config config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app") // app.env
	viper.SetConfigType("env") // app.env; could also use JSON, xml, etc.

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	// Unmarshal: parsing serialized data (e.g., JSON, XML) into a Go data structure for use in the program.
	return
}

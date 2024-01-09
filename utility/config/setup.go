package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	server_port   int
	MQTT_port     int
	MQTT_server   string
	MQTT_username string
	MQTT_password string
}

var GlobalConfig *Config

func Setup() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	GlobalConfig = &Config{
		server_port:   viper.GetInt("SERVER_PORT"),
		MQTT_port:     viper.GetInt("MQTT_PORT"),
		MQTT_server:   viper.GetString("MQTT_SERVER"),
		MQTT_username: viper.GetString("MQTT_USERNAME"),
		MQTT_password: viper.GetString("MQTT_PASSWORD"),
	}

}

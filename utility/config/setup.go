package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server_port    int
	MQTT_port      int
	MQTT_server    string
	MQTT_username  string
	MQTT_password  string
	Download_URL   string
	MySQL_host     string
	MySQL_port     int
	MySQL_dbname   string
	MySQL_username string
	MySQL_password string
}

var GlobalConfig *Config

func Setup() {
	viper.AutomaticEnv()
	GlobalConfig = &Config{
		Server_port:    viper.GetInt("SERVER_PORT"),
		MQTT_port:      viper.GetInt("MQTT_PORT"),
		MQTT_server:    viper.GetString("MQTT_SERVER"),
		MQTT_username:  viper.GetString("MQTT_USERNAME"),
		MQTT_password:  viper.GetString("MQTT_PASSWORD"),
		Download_URL:   viper.GetString("DOWNLOAD_URL"),
		MySQL_host:     viper.GetString("MYSQL_HOST"),
		MySQL_port:     viper.GetInt("MYSQL_PORT"),
		MySQL_dbname:   viper.GetString("MYSQL_DBNAME"),
		MySQL_username: viper.GetString("MYSQL_USER"),
		MySQL_password: viper.GetString("MYSQL_PASSWORD"),
	}

}

package constants

import (
	"github.com/spf13/viper"
	"os"
)

var CacheDuration string
var Port string

func init() {
	// todo: remove this
	os.Setenv("APP_CACHE_DURATION", "10s")
	os.Setenv("PORT", "8080")

	// system envs
	viper.BindEnv("port")

	// app specific envs
	viper.SetEnvPrefix("app")
	viper.BindEnv("cache_duration")

	CacheDuration = viper.GetString("cache_duration")
	Port = viper.GetString("port")
}

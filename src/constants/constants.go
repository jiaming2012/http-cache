package constants

import (
	"github.com/spf13/viper"
	"os"
)

var CacheDuration string
var Port string
var EsdbURL string
var HttpCacheStreamName string
var HttpCacheStreamVersion uint
var AppMode string

func init() {
	// todo: remove this
	os.Setenv("APP_CACHE_DURATION", "20s")
	os.Setenv("PORT", "8080")
	os.Setenv("APP_ESDB_URL", "esdb+discover://localhost:2113?tls=false&keepAliveTimeout=10000&keepAliveInterval=10000")
	os.Setenv("APP_HTTP_CACHE_STREAM_NAME", "http-cache-stream")
	os.Setenv("APP_MODE", "eventstoredb")
	os.Setenv("APP_HTTP_CACHE_STREAM_VERSION", "1")

	// system envs
	viper.BindEnv("port")

	// app specific envs
	viper.SetEnvPrefix("app")
	viper.BindEnv("cache_duration")
	viper.BindEnv("esdb_url")
	viper.BindEnv("http_cache_stream_name")
	viper.BindEnv("http_cache_stream_version")
	viper.BindEnv("mode")

	CacheDuration = viper.GetString("cache_duration")
	Port = viper.GetString("port")
	EsdbURL = viper.GetString("esdb_url")
	HttpCacheStreamName = viper.GetString("http_cache_stream_name")
	HttpCacheStreamVersion = viper.GetUint("http_cache_stream_version")
	AppMode = viper.GetString("mode")
}

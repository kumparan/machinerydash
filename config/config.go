package config

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	GetConf()
}

// GetConf :nodoc:
func GetConf() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./..")
	viper.AddConfigPath("./../..")
	viper.AddConfigPath("./../../..")
	viper.SetConfigName("config")
	viper.SetEnvPrefix("svc")

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Warningf("%v", err)
	}
}

// Env :nodoc:
func Env() string {
	return viper.GetString("env")
}

// LogLevel :nodoc:
func LogLevel() string {
	return viper.GetString("log_level")
}

// SentryDSN :nodoc:
func SentryDSN() string {
	return viper.GetString("sentry_dsn")
}

// Port :nodoc:
func Port() string {
	return viper.GetString("port")
}

// DynamodbHost :nodoc:
func DynamodbHost() string {
	return viper.GetString("dynamodb.host")
}

// DynamodbRegion :nodoc:
func DynamodbRegion() string {
	return viper.GetString("dynamodb.region")
}

// DynamodbTaskTable :nodoc:
func DynamodbTaskTable() string {
	return viper.GetString("dynamodb.task_table")
}

// DynamodbGroupTable :nodoc:
func DynamodbGroupTable() string {
	return viper.GetString("dynamodb.group_table")
}

// IsLocalDynamodb :nodoc:
func IsLocalDynamodb() bool {
	return viper.GetBool("dynamodb.is_local")
}

// MachineryResultExpiry :nodoc:
func MachineryResultExpiry() int {
	return viper.GetInt("machinery.result_expiry")
}

// MachineryBrokerNamespace :nodoc:
func MachineryBrokerNamespace() string {
	return viper.GetString("machinery.broker_namespace")
}

// MachineryBrokerHost :nodoc:
func MachineryBrokerHost() string {
	return viper.GetString("machinery.broker_host")
}

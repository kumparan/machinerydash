module github.com/kumparan/machinerydash

go 1.14

require (
	bou.ke/monkey v1.0.2
	github.com/RichardKnop/machinery v1.9.2
	github.com/aws/aws-sdk-go v1.33.6
	github.com/banzaicloud/logrus-runtime-formatter v0.0.0-20190729070250-5ae5475bae5e
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054 // indirect
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/kumparan/go-utils v1.7.0
	github.com/labstack/echo/v4 v4.1.17
	github.com/markbates/pkger v0.17.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
)

replace github.com/RichardKnop/machinery => github.com/kumparan/machinery v1.9.3-0.20201202083018-181992f5f0eb

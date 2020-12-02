package console

import (
	"fmt"
	"os"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/evalphobia/logrus_sentry"
	"github.com/kumparan/machinerydash/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "machinerydash",
	Short: "machinery dashboard",
	Long:  `Dashboard for machinery`,
}

// Execute :nodoc:
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolP("integration", "i", false, "use integration config file")
	config.GetConf()
	setupLogger()

}

func setupLogger() {
	formatter := runtime.Formatter{
		ChildFormatter: &logrus.JSONFormatter{},
		Line:           true,
		File:           true,
	}

	if config.Env() == "development" {
		formatter = runtime.Formatter{
			ChildFormatter: &logrus.TextFormatter{
				ForceColors:   true,
				FullTimestamp: true,
			},
			Line: true,
			File: true,
		}
	}

	logrus.SetFormatter(&formatter)
	logrus.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(config.LogLevel())
	if err != nil {
		logLevel = logrus.DebugLevel
	}
	logrus.SetLevel(logLevel)

	hook, err := logrus_sentry.NewSentryHook(config.SentryDSN(), []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	if err != nil {
		logrus.Info("Logger configured to use only local stdout")
		return
	}

	hook.SetEnvironment(config.Env())
	hook.Timeout = 0 // fire and forget
	hook.StacktraceConfiguration.Enable = true
	logrus.AddHook(hook)
}

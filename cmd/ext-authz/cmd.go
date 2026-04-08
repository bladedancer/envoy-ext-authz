package cmd

import (
	"strings"

	extauthz "github.com/bladedancer/envoy-ext-authz/pkg/ext-authz"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:     "extauthzdemo",
	Short:   "Envoy ext_authz demo with jwt_authn metadata forwarding.",
	Version: "0.0.1",
	RunE:    run,
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.Flags().Uint32("port", 10001, "The gRPC port to listen on.")
	RootCmd.Flags().String("logLevel", "info", "log level")
	RootCmd.Flags().String("logFormat", "json", "line or json")

	bindOrPanic("port", RootCmd.Flags().Lookup("port"))
	bindOrPanic("log.level", RootCmd.Flags().Lookup("logLevel"))
	bindOrPanic("log.format", RootCmd.Flags().Lookup("logFormat"))
}

func initConfig() {
	viper.SetTypeByDefaultValue(true)
	viper.SetEnvPrefix("extauthz")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func bindOrPanic(key string, flag *flag.Flag) {
	if err := viper.BindPFlag(key, flag); err != nil {
		panic(err)
	}
}

func run(cmd *cobra.Command, args []string) error {
	logger, err := setupLogging(viper.GetString("log.level"), viper.GetString("log.format"))
	if err != nil {
		return err
	}

	extauthz.Init(logger, extauthzConfig())
	return extauthz.Run()
}

func extauthzConfig() *extauthz.Config {
	return &extauthz.Config{
		Port: viper.GetUint32("port"),
	}
}

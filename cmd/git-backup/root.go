/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stenic/go-git-backup/pkg/app"
)

var v string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "git-backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(cmd.Context(), app.Platform{
			Name:         "github",
			Organisation: viper.GetString("github.organisation"),
		})
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return setUpLogs(os.Stdout, v)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	viper.SetDefault("log.level", "info")
	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", viper.GetString("log.level"), "Log level (debug, info, warn, error, fatal, panic")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	Execute()
}

func setUpLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}

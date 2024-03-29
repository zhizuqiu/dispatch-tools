// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	address      string
	dir          string
	httpUser     string
	httpPassword string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dispatch",
	Short: "dispatch tools",
	Long:  `dispatch tools`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//    Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件 (默认：$HOME/.dispatch/dispatch.yaml)")

	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "http://127.0.0.1:8080/", "dispatch server 的地址，例如：-a http://127.0.0.1:8080/")
	rootCmd.PersistentFlags().StringVarP(&dir, "dir", "d", "/", "dispatch server 的目录路径，例如：-d /temp/")
	rootCmd.PersistentFlags().StringVar(&httpUser, "http-user", "", "dispatch server 的认证用户")
	rootCmd.PersistentFlags().StringVar(&httpPassword, "http-password", "", "dispatch server 的认证密码")

	// Note: the variable address will not be set to the value from config, when the --address flag is not provided by user.
	_ = viper.BindPFlag("address", rootCmd.PersistentFlags().Lookup("address"))
	_ = viper.BindPFlag("dir", rootCmd.PersistentFlags().Lookup("dir"))
	_ = viper.BindPFlag("http-user", rootCmd.PersistentFlags().Lookup("http-user"))
	_ = viper.BindPFlag("http-password", rootCmd.PersistentFlags().Lookup("http-password"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".demo" (without extension).
		viper.AddConfigPath(home + "/.dispatch/")
		viper.SetConfigName("dispatch")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

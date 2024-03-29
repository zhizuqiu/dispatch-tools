// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"dispatch/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Args:  cobra.ExactArgs(1),
	Short: "上传文件",
	Long: `上传文件

用法：
dispatch up ./some.zip
dispatch up -a http://127.0.0.1:8080/ ./some.zip
dispatch up -a http://127.0.0.1:8080/ -d /temp/ ./some.zip
`,
	Run: func(cmd *cobra.Command, args []string) {
		confAddress := viper.GetString("address")
		confDir := viper.GetString("dir")
		confUser := viper.GetString("http-user")
		confPassword := viper.GetString("http-password")

		service.Upload(confAddress, confDir, confUser, confPassword, args[0])
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}

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

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"dl"},
	Args:    cobra.ExactArgs(1),
	Short:   "下载文件",
	Long: `下载文件

用法：
dispatch download http://127.0.0.1/some.zip
dispatch download -p /temp/ http://127.0.0.1/some.zip
`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		confUser := viper.GetString("user")
		confPassword := viper.GetString("password")
		service.Download(confUser, confPassword, path, args[0])
	},
}

func init() {
	downloadCmd.Flags().StringP("path", "p", "./", "要下载到的目录路径")
	rootCmd.AddCommand(downloadCmd)
}

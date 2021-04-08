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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    NoArgs,
	Short:   "查询文件",
	Long: `查询文件

用法：
dispatch list
dispatch list -a http://127.0.0.1:8080/
dispatch list -a http://127.0.0.1:8080/ -d /temp/
`,
	Run: func(cmd *cobra.Command, args []string) {
		confAddress := viper.GetString("address")
		confDir := viper.GetString("dir")
		wide, _ := cmd.Flags().GetBool("wide")

		service.List(confAddress, confDir, wide)
	},
}

func init() {
	listCmd.Flags().BoolP("wide", "w", false, "更详细的展示")
	rootCmd.AddCommand(listCmd)
}

// NoArgs returns an error if any args are included.
func NoArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.Errorf(
			"%q accepts no arguments\n\nUsage:  %s",
			cmd.CommandPath(),
			cmd.UseLine(),
		)
	}
	return nil
}

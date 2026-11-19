package cmd

import (
	"dispatch/service"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"srv"},
	Args:    cobra.NoArgs,
	Short:   "启动文件服务端",
	Long: `启动文件服务端，供 dispatch up / list / download 使用

用法：
dispatch server
dispatch server -d /data
dispatch server -d /data --listen 0.0.0.0:8080
dispatch server -d /data --listen 0.0.0.0:8080 --http-user admin --http-password 123456
`,
	Run: func(cmd *cobra.Command, args []string) {
		root := viper.GetString("dir")
		if !cmd.Root().PersistentFlags().Changed("dir") {
			root = "./data"
		}
		if root == "/" {
			fmt.Println("警告：数据目录为文件系统根目录 /，主机上的所有文件都可能被访问，请确认这是你要的。")
		}

		cfg := service.ServerConfig{
			Root:        root,
			Listen:      viper.GetString("listen"),
			User:        viper.GetString("http-user"),
			Password:    viper.GetString("http-password"),
			NoOverwrite: viper.GetBool("no-overwrite"),
		}

		if err := service.RunServer(cfg); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	serverCmd.Flags().String("listen", "127.0.0.1:8080", "服务端监听地址，例如：--listen 0.0.0.0:8080")
	serverCmd.Flags().Bool("no-overwrite", false, "上传时同名文件已存在则返回 409，不覆盖")
	_ = viper.BindPFlag("listen", serverCmd.Flags().Lookup("listen"))
	_ = viper.BindPFlag("no-overwrite", serverCmd.Flags().Lookup("no-overwrite"))
	rootCmd.AddCommand(serverCmd)
}

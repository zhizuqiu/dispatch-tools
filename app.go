package main

import (
	"dispatch-up/service"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"strings"
)

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

var (
	config  = flag.String("config", getEnv("DISPATCH_CONFIG", ""), "dispatch配置文件路径")
	address = flag.String("a", getEnv("DISPATCH_ADDRESS", ""), "dispatch server地址，例如：http://127.0.0.1:8080/")
	dir     = flag.String("d", getEnv("DISPATCH_DIR", ""), "要上传到的目录路径，例如：/temp/")
	file    = flag.String("f", getEnv("DISPATCH_DIR", ""), "要上传的文件路径，例如：./nginx.conf")
	h       = flag.Bool("h", false, "show help.")
)

func usage() {
	flag.Usage()
	fmt.Println("例如：")
	fmt.Println("dispatch-up -f ./some.zip")
	fmt.Println("dispatch-up -a http://127.0.0.1:8080/ -f ./some.zip")
	fmt.Println("dispatch-up -a http://127.0.0.1:8080/ -d /temp/ -f ./some.zip")
}

func main() {
	flag.Parse()
	if *h {
		usage()
		return
	}

	if *config == "" {
		config = stringP("./config")
		currentUser, err := user.Current()
		if nil == err {
			config = stringP(currentUser.HomeDir + "/.dispatch/config")
		}
	}

	setting := service.InitConfigFromFile(*config)
	if *address != "" {
		setting.Address = *address
	}
	if *dir != "" {
		setting.Dir = *dir
	}

	u, err := url.Parse(setting.Address)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if setting.Dir == "" {
		setting.Dir = "/"
	}

	if setting.Dir != "/" {
		if !strings.HasPrefix(setting.Dir, "/") {
			setting.Dir = "/" + setting.Dir
		}
		if !strings.HasSuffix(setting.Dir, "/") {
			setting.Dir = setting.Dir + "/"
		}
	}

	fmt.Println("address: " + setting.Address)
	fmt.Println("dir: " + setting.Dir)
	fmt.Println("file: " + *file)

	if setting.Address == "" || *file == "" {
		usage()
		return
	}

	service.Upload(u.Scheme+"://"+u.Host+"/file/UploadFile?Path="+setting.Dir, *file)
}

func stringP(s string) *string {
	return &s
}

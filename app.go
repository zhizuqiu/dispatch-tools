package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	ruisUtil "github.com/mgr9525/go-ruisutil"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

var (
	address = flag.String("a", getEnv("DISPATCH_ADDRESS", ""), "dispatch server地址，例如：127.0.0.1:8080")
	dir     = flag.String("d", getEnv("DISPATCH_DIR", "/"), "要上传到的目录路径，例如：/temp/")
	file    = flag.String("f", getEnv("DISPATCH_DIR", ""), "要上传的文件路径，例如：./nginx.conf")
)

func main() {
	flag.Parse()

	u, err := url.Parse(*address)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if *dir == "" {
		dir = stringP("/")
	}

	if *dir != "/" {
		if !strings.HasPrefix(*dir, "/") {
			dir = stringP("/" + *dir)
		}
		if !strings.HasSuffix(*dir, "/") {
			dir = stringP(*dir + "/")
		}
	}

	if *address == "" || *file == "" {
		flag.Usage()
	}

	upload("http://"+u.Host+"/file/UploadFile?Path="+*dir, *file)
}

func stringP(s string) *string {
	return &s
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func upload(url, flpath string) {
	body := ruisUtil.NewCircleByteBuffer(10240)
	boundary := randomBoundary()
	boundarybytes := []byte("\r\n--" + boundary + "\r\n")
	endbytes := []byte("\r\n--" + boundary + "--\r\n")

	reqest, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Content-Type", "multipart/form-data; charset=utf-8; boundary="+boundary)
	go func() {
		f, err := os.OpenFile(flpath, os.O_RDONLY, 0666)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		stat, err := f.Stat() //获取文件状态
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer f.Close()

		header := fmt.Sprintf("Content-Disposition: form-data; name=\"upfile\"; filename=\"%s\"\r\nContent-Type: application/octet-stream\r\n\r\n", stat.Name())
		body.Write(boundarybytes)
		body.Write([]byte(header))

		fsz := float64(stat.Size())
		fupsz := float64(0)
		buf := make([]byte, 1024)
		for {
			n, err := f.Read(buf)
			if n > 0 {
				nz, _ := body.Write(buf[0:n])
				fupsz += float64(nz)
				i := int((fupsz / fsz) * 100)
				progress := strconv.Itoa(i) + "%"
				j := (100 - i) / 2
				h := strings.Repeat("=", 50-j) + strings.Repeat(" ", j)
				fmt.Print("\r["+h+"] | ", progress, " | ", strconv.FormatFloat(fupsz, 'f', 0, 64), "/", stat.Size())
			}
			if err == io.EOF {
				break
			}
		}
		body.Write(endbytes)
		body.Write(nil)
	}()
	resp, err := http.DefaultClient.Do(reqest)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("\n上传成功")
	} else {
		fmt.Println("\n上传失败,StatusCode:", resp.StatusCode)
	}
}

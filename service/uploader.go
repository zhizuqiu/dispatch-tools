package service

import (
	"crypto/rand"
	"fmt"
	ruisUtil "github.com/mgr9525/go-ruisutil"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	HttpClientNoTimeout = &http.Client{
		Timeout: 0,
	}
)

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func Upload(address, dir, file string) {

	dir = parseDir(dir)
	_url, err := getUploadUrl(address, dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("address: " + address)
	fmt.Println("dir: " + dir)
	fmt.Println("file: " + file)
	fmt.Println("")

	if address == "" || file == "" {
		return
	}

	body := ruisUtil.NewCircleByteBuffer(10240)
	boundary := randomBoundary()
	boundarybytes := []byte("\r\n--" + boundary + "\r\n")
	endbytes := []byte("\r\n--" + boundary + "--\r\n")

	reqest, err := http.NewRequest("POST", _url, body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Content-Type", "multipart/form-data; charset=utf-8; boundary="+boundary)

	go func() {
		f, err := os.OpenFile(file, os.O_RDONLY, 0666)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		stat, err := f.Stat()
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
				nz, err := body.Write(buf[0:n])
				if err != nil {
					fmt.Println(err)
				}
				fupsz += float64(nz)
				i := int((fupsz / fsz) * 100)
				progress := strconv.Itoa(i) + "%"
				j := (100 - i) / 2
				h := strings.Repeat("=", 50-j) + strings.Repeat(" ", j)
				fmt.Print("\r["+h+"]  ", progress, "  ", strconv.FormatFloat(fupsz, 'f', 0, 64), "/", stat.Size())
			}
			if err == io.EOF {
				break
			}
		}
		body.Write(endbytes)
		body.Write(nil)
	}()

	resp, err := HttpClientNoTimeout.Do(reqest)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("\nsuccessful uploaded!")
	} else {
		fmt.Println("\nupload failed! code:", resp.StatusCode)
	}
}

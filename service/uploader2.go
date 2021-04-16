package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// 不带进度条
func Upload2(address, dir, file string) {

	dir = parseDir(dir)
	_url, err := getUploadUrl(address, dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println()
	fmt.Println("address: " + address)
	fmt.Println("dir: " + dir)
	fmt.Println("file: " + file)
	fmt.Println()

	if address == "" || file == "" {
		return
	}

	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("upfile", filepath.Base(f.Name()))
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = io.Copy(part, f)
	if err != nil {
		fmt.Println(err)
		return
	}
	writer.Close()

	request, err := http.NewRequest("POST", _url, body)
	if err != nil {
		fmt.Println(err)
		return
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		fmt.Println("\nUpload Completed!")
	} else {
		fmt.Println("\nUpload Failed! code:", response.StatusCode)
	}
}

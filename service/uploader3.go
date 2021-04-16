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

func Upload3(address, dir, file string) {

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

	/*
		stat, err := f.Stat()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	*/

	/*
		reader := &Reader{
			Reader: f,
			Total:  stat.Size(),
			Bar:    pb.ProgressBarTemplate(downloadBarTmpl).Start64(stat.Size()),
		}
		reader.Bar.SetMaxWidth(100)
	*/

	part, err := writer.CreateFormFile("upfile", filepath.Base(f.Name()))
	if err != nil {
		fmt.Println(err)
		return
	}

	/*
		w := &Writer{
			Writer: part,
			Total:  stat.Size(),
			Bar:    pb.ProgressBarTemplate(downloadBarTmpl).Start64(stat.Size()),
		}
		w.Bar.SetMaxWidth(100)
	*/

	_, err = io.Copy(part, f)
	if err != nil {
		fmt.Println(err)
		return
	}
	writer.Close()

	// w.Bar.Finish()

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

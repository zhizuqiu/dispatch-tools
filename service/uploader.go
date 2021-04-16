package service

import (
	"bufio"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
)

func Upload(address, dir, file string) {
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

	response, err := uploadFileMultipart(_url, file)
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

func uploadFileMultipart(url string, path string) (*http.Response, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Reduce number of syscalls when reading from disk.
	bufferedFileReader := bufio.NewReader(f)
	defer f.Close()

	// Create a pipe for writing from the file and reading to
	// the request concurrently.
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)

	// Store the first write error in writeErr.
	var (
		writeErr error
		errOnce  sync.Once
	)
	setErr := func(err error) {
		if err != nil {
			errOnce.Do(func() { writeErr = err })
		}
	}

	stat, err := f.Stat()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	reader := &Reader{
		Reader: bufferedFileReader,
		Total:  stat.Size(),
		Bar:    pb.ProgressBarTemplate(uploadBarTmpl).Start64(stat.Size()),
	}
	reader.Bar.SetMaxWidth(100)

	go func() {
		partWriter, err := formWriter.CreateFormFile("upfile", path)
		setErr(err)
		_, err = io.Copy(partWriter, reader)
		setErr(err)
		setErr(formWriter.Close())
		setErr(bodyWriter.Close())
		reader.Bar.Finish()
	}()

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())

	// This operation will block until both the formWriter
	// and bodyWriter have been closed by the goroutine,
	// or in the event of a HTTP error.
	resp, err := http.DefaultClient.Do(req)

	if writeErr != nil {
		return nil, writeErr
	}

	return resp, err
}

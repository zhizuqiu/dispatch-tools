package service

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

func Download(dpath, durl string) {

	dpath = getDownloadPath(dpath)

	uri, err := url.ParseRequestURI(durl)
	if err != nil {
		fmt.Println(err)
		return
	}

	filename := path.Base(uri.Path)

	res, err := http.Get(durl)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = res.Body.Close() }()

	filepath := dpath + filename

	_, err = os.Stat(filepath)
	if !os.IsNotExist(err) {
		var yes string
		fmt.Print("file [" + filepath + "] already exists. Overwrite (y/n):")
		_, err = fmt.Scan(&yes)
		if err != nil {
			fmt.Println(err)
			return
		}
		if yes != "y" {
			return
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = f.Close() }()

	reader := &Reader{
		Reader: res.Body,
		Total:  res.ContentLength,
		Bar:    pb.ProgressBarTemplate(downloadBarTmpl).Start64(res.ContentLength),
	}
	reader.Bar.SetMaxWidth(100)

	fmt.Println()
	fmt.Println("path: " + dpath)
	fmt.Println("url: " + durl)
	fmt.Println("file path: " + filepath)
	fmt.Println()

	_, err = io.Copy(f, reader)
	if err != nil {
		fmt.Println(err)
		return
	}
	reader.Bar.Finish()

	fmt.Println("\nDownload Successful!")
}

type Reader struct {
	io.Reader
	Total   int64
	Current int64
	Bar     *pb.ProgressBar
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)

	r.Current += int64(n)
	r.Bar.SetCurrent(r.Current)
	return
}

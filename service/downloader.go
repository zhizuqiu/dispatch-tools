package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
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
	}

	fmt.Println("path: " + dpath)
	fmt.Println("url: " + durl)
	fmt.Println("filepath: " + filepath)
	fmt.Println("")

	_, err = io.Copy(f, reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("\ndownload successful!")
}

type Reader struct {
	io.Reader
	Total   int64
	Current int64
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)

	r.Current += int64(n)
	// fmt.Printf("\r进度 %.2f%%", float64(r.Current*10000/r.Total)/100)
	i := int(r.Current * 100 / r.Total)
	progress := strconv.Itoa(i) + "%"
	j := (100 - i) / 2
	h := strings.Repeat("=", 50-j) + strings.Repeat(" ", j)
	fmt.Print("\r["+h+"]  ", progress, "  ", strconv.FormatFloat(float64(r.Current), 'f', 0, 64), "/", r.Total)
	return
}

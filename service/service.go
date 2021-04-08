package service

import (
	"net/url"
	"strings"
)

func parseDir(dir string) string {
	if dir == "" {
		dir = "/"
	}

	if dir != "/" {
		if !strings.HasPrefix(dir, "/") {
			dir = "/" + dir
		}
		if !strings.HasSuffix(dir, "/") {
			dir = dir + "/"
		}
	}

	return dir
}

func getUploadUrl(address, dir string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}

	return u.Scheme + "://" + u.Host + "/file/UploadFile?Path=" + dir, nil
}

func getListUrl(address, dir string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}

	return u.Scheme + "://" + u.Host + "/file/List?Path=" + dir, nil
}

func getDownloadUrl(address, path string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}

	return u.Scheme + "://" + u.Host + "/data" + path, nil
}

func getDownloadPath(path string) string {
	if path == "." {
		path = "./"
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return path
}

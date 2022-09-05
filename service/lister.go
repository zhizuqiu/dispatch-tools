package service

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var (
	HttpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*20)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
)

type PathSlice []*Path

// 按名字排序
func (s PathSlice) Len() int      { return len(s) }
func (s PathSlice) Swap(i, j int) { *s[i], *s[j] = *s[j], *s[i] }
func (s PathSlice) Less(i, j int) bool {
	if (*s[i]).IsDir && !(*s[j]).IsDir {
		return true
	}
	if !(*s[i]).IsDir && (*s[j]).IsDir {
		return false
	}
	return (*s[i]).Name < (*s[j]).Name
}

type Path struct {
	Id           int
	IsDir        bool
	LastModified string
	Length       int64
	Name         string
	Path         string
}

func List(address, dir, user, password string, wide bool) {

	dir = parseDir(dir)
	_url, err := getListUrl(address, dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("address: " + address)
	fmt.Println("dir: " + dir)
	fmt.Println("")

	if address == "" {
		return
	}

	req, err := http.NewRequest(http.MethodGet, _url, http.NoBody)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.SetBasicAuth(user, password)

	resp, err := HttpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var pathSlice PathSlice
	err = json.Unmarshal(body, &pathSlice)
	if err != nil {
		fmt.Println(err)
		return
	}

	table := uitable.New()
	// table.MaxColWidth = 50
	table.Separator = "  "

	if wide {
		table.AddRow("Name", "LastModified", "Length", "Url")
		for _, f := range pathSlice {
			url, err := getDownloadUrl(address, f.Path)
			if err != nil {
				fmt.Println(err)
				return
			}
			if f.IsDir {
				table.AddRow(color.HiBlueString(f.Name), f.LastModified, f.Length, url)
			} else {
				table.AddRow(f.Name, f.LastModified, f.Length, url)
			}
		}
	} else {
		table.AddRow("Name", "LastModified", "Length")
		for _, f := range pathSlice {
			if f.IsDir {
				table.AddRow(color.HiBlueString(f.Name), f.LastModified, f.Length)
			} else {
				table.AddRow(f.Name, f.LastModified, f.Length)
			}
		}
	}
	fmt.Println(table)
}

package main

import (
	"net/url"
	"os"

	"path/filepath"

	"net/http"

	"fmt"

	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/antchfx/xquery/html"
)

func main() {
	if len(os.Args) != 2 {
		panic("Must input url as the only argument")
	}

	strUrl := os.Args[1]
	uri, err := url.Parse(strUrl)
	if err != nil {
		panic(err)
	}

	fileId := ""
	if uri.Query().Get("id") != "" {
		fileId = uri.Query().Get("id")
	} else {
		fileId = filepath.Base(filepath.Dir(uri.Path))
	}

	if len(fileId) < 20 {
		log.Error("FileId not parsed correctly for: " + strUrl)
		os.Exit(1)
	}

	resp, err := http.Get(fmt.Sprintf("https://docs.google.com/uc?id=%s&export=download", fileId))
	if err != nil {
		panic(err)
	}
	cookies := resp.Cookies()
	defer resp.Body.Close()

	root, err := htmlquery.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	filenameNode := htmlquery.FindOne(root, "//span[@class='uc-name-size']/a")
	filename := filenameNode.FirstChild.Data

	downloadLinkNode := htmlquery.FindOne(root, "//*[@id='uc-download-link']/@href")

	downloadPath := ""
	for _, attr := range downloadLinkNode.Attr {
		if attr.Key == "href" {
			downloadPath = "https://docs.google.com" + attr.Val
			break
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("GET", downloadPath, nil)
	if err != nil {
		panic(err)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	dlResp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer dlResp.Body.Close()

	_, err = io.Copy(file, dlResp.Body)
	if err != nil {
		panic(err)
	}

	log.Infof("%s downloaded", filename)
}

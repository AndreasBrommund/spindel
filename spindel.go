package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

func main() {
	url := "http://www.wallstedts.se"
	//url = "http://www.brommund.se"
	html := DownloadPage(url)

	fmt.Printf("%s\n", html)

	links := GetLinks(string(html))

	for _, link := range links {
		fmt.Println(link)
	}
}

func DownloadPage(url string) string {
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	return string(html)
}

func GetLinks(content string) []string {
	hrefRegexp := regexp.MustCompile(`(?mUis)href="(.*)"`)
	links := make([]string, 0)

	for _, link := range hrefRegexp.FindAllStringSubmatch(content, -1) {
		links = append(links, link[1])
	}

	return links
}

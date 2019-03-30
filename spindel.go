package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Visited struct {
	v   map[string]bool
	mux sync.Mutex
}

func (v *Visited) Add(url string) {
	v.mux.Lock()
	v.v[url] = true
	v.mux.Unlock()
}

func (v *Visited) Exists(url string) bool {
	v.mux.Lock()
	defer v.mux.Unlock()
	return v.v[url]
}

var visited = Visited{v: make(map[string]bool, 0)}

var baseURL = url.URL{}

const debugLevel = 3

var wg sync.WaitGroup

func main() {
	//b, _ := url.Parse("http://www.wallstedts.se")
	//b, _ := url.Parse("http://www.brommund.se")
	b, _ := url.Parse("http://www.aftonbladet.se")

	baseURL = *b
	VisitPage(baseURL)

	wg.Wait()
	fmt.Println("Done")
}

func VisitPage(url url.URL) {
	fmt.Println("Visiting: " + url.String())
	visited.Add(url.String())
	html := DownloadPage(url.String())
	links := GetLinks(string(html))
	if len(links) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			VisitPages(links)
		}()
	}
}

func VisitPages(urls []url.URL) {
	for _, u := range urls {
		debug(fmt.Sprintf("%s = %v", u.String(), shouldVisit(u)), 5)
		if shouldVisit(u) {
			VisitPage(u)
		}
	}
}

func DownloadPage(url string) string {

	resp, err := http.Get(url)

	if err != nil {
		debug("Could not get page: "+url+" ("+err.Error()+")", 2)
		return ""
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		debug("Could not get page "+url+" ("+strconv.Itoa(resp.StatusCode)+")", 2)
		return ""
	}
	html, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		debug("Could not read response body ("+err.Error()+")", 2)
		return ""
	}

	return string(html)
}

func GetLinks(content string) []url.URL {
	hrefRegexp := regexp.MustCompile(`(?mUis)href="(.*)"`) // Fix js src=""
	urls := make([]url.URL, 0)

	for _, link := range hrefRegexp.FindAllStringSubmatch(content, -1) {
		if url, err := url.Parse(link[1]); err == nil {
			if url.Host == "" && url.Scheme == "" {
				url.Host = baseURL.Host
				url.Scheme = baseURL.Scheme
			}
			urls = append(urls, *url)
		} else {
			debug("Could not parse: "+link[1]+" ("+err.Error()+")", 2)
		}
	}

	return urls
}

func shouldVisit(url url.URL) bool {

	if visited.Exists(url.String()) {
		debug(url.String()+" already visited", 4)
		return false
	}

	if len(url.Path) > 0 {
		if strings.Contains(url.Path, ".") {
			reg := regexp.MustCompile(`(?mi).*(html|css|js|php)$`) //TOOD improve
			if !reg.MatchString(url.Path) {
				debug(url.String()+" do not match regex ("+url.Path+")", 4)
				return false
			}
		}
	}

	if !(url.Scheme == "http" || url.Scheme == "https") {
		debug(url.String()+" not https or http ("+url.Scheme+")", 4)
		return false
	}

	if url.Host != baseURL.Host {
		debug(url.String()+" only visit links at from base url host ("+url.Host+")", 4)
		return false
	}

	return true
}

func debug(msg string, level int) {
	/*
	* 1 FATAL
	* 2 ERROR
	* 3 WARN
	* 4 INFO
	* 5 DEBUG
	 */

	if debugLevel >= level {
		log.Println(msg + " (" + strconv.Itoa(level) + ")")
	}
}

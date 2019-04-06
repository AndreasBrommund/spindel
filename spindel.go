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
	"time"
)

type Visited struct {
	v   map[string]bool
	mux sync.Mutex
}

func (v *Visited) Visit(url string) bool {
	v.mux.Lock()
	defer v.mux.Unlock()

	isVisited := v.v[url]
	v.v[url] = true

	return isVisited
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

	start := time.Now()

	fetchingWorkers := 10
	parsingWorkers := 10
	buffer := 1000000 //TODO is it possible to prevent "deadlock" if the buffer is too small?

	var waitGroupFetching sync.WaitGroup
	waitGroupFetching.Add(fetchingWorkers)

	var waitGroupParsing sync.WaitGroup
	waitGroupParsing.Add(parsingWorkers)

	fetchChannel := make(chan url.URL, buffer)
	parseChannel := make(chan string, buffer)

	wg.Add(1)
	fetchChannel <- baseURL

	for w := 1; w <= fetchingWorkers; w++ {
		go func(id int) {
			defer waitGroupFetching.Done()
			fetch(id, fetchChannel, parseChannel)
		}(w)
	}

	for w := 1; w <= parsingWorkers; w++ {
		go func(id int) {
			defer waitGroupParsing.Done()
			parse(id, fetchChannel, parseChannel)
		}(w)
	}

	wg.Wait()
	close(fetchChannel)
	close(parseChannel)
	waitGroupFetching.Wait()
	waitGroupParsing.Wait()
	fmt.Println("Done")
	fmt.Println(time.Since(start))

}

func fetch(id int, fetchChannel <-chan url.URL, parseChannel chan<- string) {
	for u := range fetchChannel {
		debug(strconv.Itoa(id)+" is working on "+u.String(), 4)
		fmt.Println("Visiting: ", u.String())
		wg.Add(1)
		parseChannel <- DownloadPage(u.String())
		wg.Done()
	}
	debug("Done "+strconv.Itoa(id), 4)
}

func parse(id int, fetchChannel chan<- url.URL, parseChannel <-chan string) {
	for html := range parseChannel {
		debug(strconv.Itoa(id)+" is working", 4)
		urls := GetLinks(string(html))
		for _, u := range urls {
			if shouldVisit(u) {
				wg.Add(1)
				fetchChannel <- u
			}
		}
		wg.Done()
	}
	debug("Done "+strconv.Itoa(id), 4)
}

func DownloadPage(url string) string {
	//TODO Better error handeling  die in a good way

	var err error
	var resp *http.Response

	for i := 0; i < 10; i++ {
		resp, err = http.Get(url)
		if err != nil {
			debug("Could not get page: "+url+" ("+err.Error()+")", 2)
			return ""
		} else {
			defer resp.Body.Close()
			break
		}
	}

	if resp.StatusCode != http.StatusOK {
		debug("Could not get page "+url+" ("+strconv.Itoa(resp.StatusCode)+")", 2)
	}
	html, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		debug("Could not read response body ("+err.Error()+")", 2)
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

	if visited.Visit(url.String()) {
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
		debug(url.String()+" only visit links from base url host ("+url.Host+")", 4)
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
		log.Println(msg, " (", level, ")")
	}
}

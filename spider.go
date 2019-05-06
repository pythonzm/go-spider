package main

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MyRoundTripper struct {
	rt http.RoundTripper
}

func (mrt MyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.108 Safari/537.36")
	return mrt.rt.RoundTrip(r)
}

func GetMatches(url string, expr string) [][]string {
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: MyRoundTripper{rt: http.DefaultTransport},
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Get %s failed, err: %s\n", url, err)
		return nil
	}

	defer func() {
		if e := resp.Body.Close(); e != nil {
			fmt.Printf("body close failed,err: %s\n", e)
			return
		}
	}()

	bytes, _ := ioutil.ReadAll(resp.Body)
	compile, _ := regexp.Compile(expr)
	matches := compile.FindAllStringSubmatch(string(bytes), -1)

	return matches
}

func GetDetailUrl(url string) []string {

	var detailUrls []string
	matches := GetMatches(url, `</span>(.{0,5})<a href="(.*?)"><span>`)
	for _, value := range matches {
		detailUrls = append(detailUrls, value[2])
	}
	detailUrls = append(detailUrls, url)

	return detailUrls
}

func ParseHomePage(baseUrl string, pageNum int, detailUrlChan chan []string) []string {
	var detailUrls []string
	for i := 1; i <= pageNum; i++ {
		go func(num int) {
			url := baseUrl + strconv.Itoa(num) + "/"
			fmt.Printf("开始解析第%d页\n", num)
			matches := GetMatches(url, `<h2><a href="(.*?)"`)
			for _, value := range matches {
				detailUrls = append(detailUrls, GetDetailUrl(value[1])...)
			}
			fmt.Printf("解析第%d页完成\n", num)
			detailUrlChan <- detailUrls
		}(i)
	}
	return detailUrls
}

func ParsePageDetail(url string, gifChan chan map[string]string) {
	matches := GetMatches(url, `<p><img(.*?)src="(.*?)"\s+alt="(.*?)"`)

	for _, value := range matches {
		if len(value[2]) == 0 {
			fmt.Printf("%s 获取gifurl失败\n", url)
		} else {
			gifChan <- map[string]string{"url": value[2], "title": value[3]}
		}
	}
}

func DownloadGif(gifInfo map[string]string, wg *sync.WaitGroup, filePath string) {

	fmt.Printf("start download %s\n", gifInfo["url"])
	resp, err := http.Get(gifInfo["url"])
	if err != nil {
		fmt.Printf("Get %s failed,err: %s", gifInfo["url"], err)
		return
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			fmt.Printf("body close failed,err: %s", e)
			return
		}
	}()

	bytes, _ := ioutil.ReadAll(resp.Body)

	var fileName string
	title := strings.Replace(gifInfo["title"], "https://", "", -1)

	if strings.HasSuffix(gifInfo["title"], ".gif") {
		fileName = filePath + title
	} else {
		fileName = filePath + title + ".gif"
	}

	if err := ioutil.WriteFile(fileName, bytes, 0664); err != nil {
		fmt.Printf("download %s failed, err: %s", gifInfo["url"], err)
		return
	}
	fmt.Printf("download %s successful\n", gifInfo["url"])
	wg.Done()
}

func Spider(num int, page int, path string) {
	wg := new(sync.WaitGroup)
	detailUrlChan := make(chan []string)
	gifChan := make(chan map[string]string, num)
	baseUrl := "https://www.8mfh.com/gifchuchu/page/"

	go ParseHomePage(baseUrl, page, detailUrlChan)

	if runtime.GOOS == "windows" && !strings.HasSuffix(path, `\`) {
		path = path + `\`
	} else if runtime.GOOS == "linux" && !strings.HasSuffix(path, `/`) {
		path = path + `/`
	}

	_, e := os.Stat(path)
	if e != nil {
		if e := os.MkdirAll(path, os.ModeDir); e != nil {
			panic("创建目录" + path + "失败!")
		}
	}

	for i := 0; i < page; i++ {
		urls := <-detailUrlChan
		for _, url := range urls {
			go ParsePageDetail(url, gifChan)
		}

		wg.Add(len(urls))
		for i := 0; i < len(urls); i++ {
			gifInfo := <-gifChan
			go DownloadGif(gifInfo, wg, path)
		}
	}

	wg.Wait()
}

func main() {
	var path string

	if runtime.GOOS == "windows" {
		path = `D:\gifs\`
	} else {
		path = `/tmp/gifs/`
	}
	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Usage = "爬取某动图网站的出处动图"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "num,n",
			Value: 10,
			Usage: "设置默认下载并发数量，默认是10",
		},
		cli.IntFlag{
			Name:  "page,p",
			Value: 2,
			Usage: "设置下载的页数，默认是2",
		},
		cli.StringFlag{
			Name:  "to",
			Value: path,
			Usage: "设置下载路径，windows默认路径是D:\\gifs,其他系统默认是/tmp/gifs `PATH`",
		},
	}

	app.Action = func(c *cli.Context) error {
		Spider(c.Int("num"), c.Int("page"), c.String("to"))
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}
}


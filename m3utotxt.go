package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

type tver struct {
	Name  string
	Group string
	Urls  string
}
type m3uer struct {
	Content []string
	Groups  []string
	Tvlist  map[string]tver
}

var (
	m3uPattern     = regexp.MustCompile("^#EXTM3U")
	m3uLinePattern = regexp.MustCompile("^#EXTINF")
	m3uLineUrl     = regexp.MustCompile("^http")
	m3uLineGroup   = regexp.MustCompile(`group\-title=['"]*([^,^"^']+)['"]*`)
	flagFilePath   string
	flagUrlPath    string
)

func init() {
	flag.StringVar(&flagFilePath, "file", "", "从文件进行转换")
	flag.StringVar(&flagUrlPath, "url", "", "从m3u网址转换")
	flag.Usage = Usage
	flag.Parse()
}
func Usage() {
	fmt.Println("支持从文件和网址转换m3u 到txt 以配合diyp\n ")
	flag.PrintDefaults()
}
func main() {
	var m3u = New()
	if flagUrlPath != "" {
		m3u.ReadUrl(flagUrlPath)
	} else if flagFilePath != "" {
		m3u.ReadFile(flagFilePath)
	} else {
		flag.PrintDefaults()
		return
	}
	result := m3u.ToText()
	fmt.Println(result)
}
func (m *m3uer) ToText() string {
	sort.Strings(m.Groups)
	textContent := strings.Builder{}
	tvlist := m.ToTvMap()
	for _, g := range m.Groups {
		title := fmt.Sprintf("%s,#genre#\n", g)
		textContent.WriteString(title)
		for _, tv := range tvlist {
			if g == tv.Group {
				textContent.WriteString(fmt.Sprintf("%s,%s\n", tv.Name, tv.Urls))
			}
		}
	}
	return textContent.String()
}

func (m *m3uer) ToTvMap() (tvlist map[string]tver) {
	var tvName string
	groups := make(map[string]bool)
	tvlist = make(map[string]tver)
	for _, lintText := range m.Content {
		isScanM3uTitle := m3uLinePattern.MatchString(lintText)
		isScanM3uUrl := m3uLineUrl.MatchString(lintText)
		if isScanM3uTitle {
			s := strings.Split(lintText, ",")
			tvName = s[1]
			tv := tver{}
			tv.Name = tvName
			maths := m3uLineGroup.FindStringSubmatch(s[0])
			// if tv is have group
			if len(maths) > 0 {
				tv.Group = maths[1]
				_, ok := groups[tvName]
				if !ok {
					groups[tv.Group] = true
				}
			}
			tvlist[tvName] = tv
		}

		if isScanM3uUrl && tvName != "" {
			tv := tvlist[tvName]
			if tv.Urls == "" {
				tv.Urls = lintText
			} else {
				tv.Urls = fmt.Sprintf("%s#@%s", tv.Urls, lintText)
			}
			tvlist[tvName] = tv
			tvName = ""
		}
	}
	for k, ok := range groups {
		_ = ok
		m.Groups = append(m.Groups, k)
	}
	m.Tvlist = tvlist
	return tvlist
}

func ioReadToSliceString(r io.Reader) []string {
	content := []string{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if !m3uPattern.MatchString(content[0]) {
		log.Fatalln("输入的不是m3u文件")
	}
	return content
}

func (m *m3uer) ReadUrl(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Panicf(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		log.Panicf("网址无效")
	}
	body := resp.Body
	defer body.Close()
	m.Content = ioReadToSliceString(body)
}
func (m *m3uer) ReadFile(filepath string) {
	f, err := os.Open(filepath)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	m.Content = ioReadToSliceString(f)
}
func New() *m3uer {
	var m m3uer
	return &m
}

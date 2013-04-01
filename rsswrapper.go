package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Rss struct {
	All     string        `xml:",innerxml"`
	Channel ChannelObject `xml:"channel"`
}

type ChannelObject struct {
	All   string       `xml:",innerxml"`
	Items []ItemObject `xml:"item"`
}

type ItemObject struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Guid        string `xml:"guid"`
	Image       string `xml:"image"`
	Author      string `xml:"author"`
	Category    string `xml:"category"`
	ParsedImage string
}

type DescriptionParser func(string) (string, string)

func AftonbladetParse(in string) (string, string) {
	var srcRegex = regexp.MustCompile(`<img [^>]*src="([^"]+)"[^>]*>`)
	var tagRegex = regexp.MustCompile(`(<([^>]+)>)`)

	in = strings.Replace(in, "<![CDATA[", "", 1)
	in = strings.Replace(in, "]]>", "", 1)
	matches := srcRegex.FindStringSubmatch(in)
	in = tagRegex.ReplaceAllString(in, "")
	if len(matches) > 1 {
		return strings.Replace(in, matches[0], "", 1), strings.Trim(matches[1], " ")
	} else {
		return in, ""
	}

	return "", ""
}

func RedditParse(in string) (string, string) {
	var imgRegex = regexp.MustCompile(`https?:\/\/(?:[a-z\-]+\.)+[a-z]{2,6}(?:\/[^\/#?]+)+\.(?:jpe?g|gif|png)`)

	matches := imgRegex.FindStringSubmatch(in)
	if len(matches) > 0 {
		return in, strings.Trim(matches[0], " ")
	} else {
		return in, ""
	}

	return "", ""
}

const timeFormat = "Mon, 2 Jan 2006 15:04:05 -0700"
const timeFormat2 = "Mon, 2 Jan 2006 15:04:05 MST"

var timeDiff = (int64)(60 * 60)

func (i ItemObject) UnixTime() int64 {
	timeString := i.PubDate
	t, err := time.Parse(timeFormat, timeString)
	if err != nil {
		t, err = time.Parse(timeFormat2, timeString)
		if err != nil {
			fmt.Println("%s for\n%s", err, i)
			return 0
		}
	}

	return t.Unix()
}

func main() {
	http.HandleFunc("/rss", rssHandler)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

//Allow sorting
type SortableItems []ItemObject

func (s SortableItems) Len() int      { return len(s) }
func (s SortableItems) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByTime struct{ SortableItems }

func (s ByTime) Less(i, j int) bool {
	return s.SortableItems[i].UnixTime() > s.SortableItems[j].UnixTime()
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	newTimeDiff := r.URL.Query().Get("time")
	if newTimeDiff != "" {
		time, err := strconv.ParseInt(newTimeDiff, 10, 64)
		if err != nil {
			fmt.Printf("%s", err)
			timeDiff = 60 * 60
		} else {
			timeDiff = time * 60 * 60
		}
	}

	channel := make(chan []ItemObject)

	feeds := 5
	go getFeed(channel, "http://www.aftonbladet.se/rss.xml", AftonbladetParse)
	go getFeed(channel, "http://www.dn.se/nyheter/m/rss/", nil)
	go getFeed(channel, "http://www.svd.se/?service=rss", nil)
	go getFeed(channel, "http://www.reddit.com/r/gifs/.rss", RedditParse)
	go getFeed(channel, "http://rss.cnn.com/rss/edition.rss", nil)
	go getFeed(channel, "http://news.google.com/?output=rss", nil)

	var items []ItemObject
	for i := 0; i < feeds; i++ {
		rec := <-channel
		for _, i := range rec {
			items = append(items, i)
		}
	}

	sort.Sort(ByTime{items})

	res, _ := json.Marshal(items)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(res))
}

func getFeed(out chan<- []ItemObject, feed string, parser DescriptionParser) {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(5 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	resp, err := c.Get(feed)

	if err != nil {
		items := make([]ItemObject, 1)
		items[0].Title = err.Error()
		out <- items
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	rss := new(Rss)
	xml.Unmarshal(body, rss)

	//Remove items older than 1 hour
	var recentItems []ItemObject
	for _, item := range rss.Channel.Items {
		if time.Now().Unix()-item.UnixTime() < timeDiff {
			recentItems = append(recentItems, item)
		}
	}

	//Parse out description and image
	items := recentItems
	if parser != nil {
		for i := 0; i < len(items); i++ {
			items[i].Description, items[i].ParsedImage = parser(items[i].Description)
		}
	}

	out <- items

}

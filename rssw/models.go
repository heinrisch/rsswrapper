package rssw

import (
	"fmt"
	"github.com/opesun/goquery"
	"strconv"
	"strings"
	"time"
)

//Types
type DescriptionParser func(out chan<- int, i *ItemObject)

//Models for RSS
type Rss struct {
	All     string        `xml:",innerxml"`
	Channel ChannelObject `xml:"channel"`
}

type ChannelObject struct {
	Title       string       `xml:"title"`
	Link        string       `xml:"link"`
	Description string       `xml:"description"`
	Items       []ItemObject `xml:"item"`
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
	Source      string
	time        int64
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

func Attr(n *goquery.Node) (string, string, int, int) {
	var src, alt string
	var height, width int
	for _, attr := range n.Attr {
		if attr.Key == "src" {
			src = attr.Val
		}
		if attr.Key == "alt" {
			alt = attr.Val
		}
		if attr.Key == "width" {
			width, _ = strconv.Atoi(strings.Replace(attr.Val, "px", "", -1))
		}
		if attr.Key == "height" {
			height, _ = strconv.Atoi(strings.Replace(attr.Val, "px", "", -1))
		}
	}

	return src, alt, width, height
}

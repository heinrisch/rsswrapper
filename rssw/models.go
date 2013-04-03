package rssw

import (
	"fmt"
	"time"
)

//Types
type DescriptionParser func(i *ItemObject)

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

//Models for HMLT
type Html struct {
	All  string     `xml:",innerxml"`
	Head HeadObject `xml:"head"`
}

type HeadObject struct {
	All  string       `xml:",innerxml"`
	Meta []MetaObject `xml:"meta"`
}

type MetaObject struct {
	All      string `xml:",innerxml"`
	Property string `xml:"property,attr"`
	Name     string `xml:"name,attr"`
	Content  string `xml:"content,attr"`
}

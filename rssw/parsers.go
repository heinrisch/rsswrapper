package rssw

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func AftonbladetParse(i *ItemObject) {
	var srcRegex = regexp.MustCompile(`<img [^>]*src="([^"]+)"[^>]*>`)
	var tagRegex = regexp.MustCompile(`(<([^>]+)>)`)

	i.Description = strings.Replace(i.Description, "<![CDATA[", "", 1)
	i.Description = strings.Replace(i.Description, "]]>", "", 1)
	matches := srcRegex.FindStringSubmatch(i.Description)
	i.Description = tagRegex.ReplaceAllString(i.Description, "")
	if len(matches) > 1 {
		i.Description = strings.Replace(i.Description, matches[0], "", 1)
		i.ParsedImage = strings.Trim(matches[1], " ")
	}
}

func RedditParse(i *ItemObject) {
	var imgRegex = regexp.MustCompile(`https?:\/\/(?:[a-z\-]+\.)+[a-z]{2,6}(?:\/[^\/#?]+)+\.(?:jpe?g|gif|png)`)

	matches := imgRegex.FindStringSubmatch(i.Description)
	if len(matches) > 0 {
		i.ParsedImage = strings.Trim(matches[0], " ")
	}
}

func SVDParse(i *ItemObject) {
	resp, err := httpGet(5, i.Link)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	html := new(Html)
	xml.Unmarshal([]byte(body), html)

	for _, meta := range html.Head.Meta {
		if meta.Property == "og:image" {
			i.ParsedImage = meta.Content
		}

	}
}

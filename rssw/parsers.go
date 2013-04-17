package rssw

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func AftonbladetParse(out chan<- int, i *ItemObject) {
	i.Description = strings.Replace(i.Description, "<![CDATA[", "", 1)
	i.Description = strings.Replace(i.Description, "]]>", "", 1)

	removeFirstImages(i)
	removeAllTags(i)

	out <- 0
}

func YahooParse(out chan<- int, i *ItemObject) {
	removeFirstImages(i)
	removeAllTags(i)

	out <- 0
}

func removeAllTags(i *ItemObject) {
	var tagRegex = regexp.MustCompile(`(<([^>]+)>)`)
	i.Description = tagRegex.ReplaceAllString(i.Description, "")
}

func removeFirstImages(i *ItemObject) bool {
	var srcRegex = regexp.MustCompile(`<img [^>]*src="([^"]+)"[^>]*>`)
	matches := srcRegex.FindStringSubmatch(i.Description)
	if len(matches) > 1 {
		i.Description = strings.Replace(i.Description, matches[0], "", 1)
		i.ParsedImage = strings.Trim(matches[1], " ")
		return true
	}

	return false
}

func RedditParse(out chan<- int, i *ItemObject) {
	var imgRegex = regexp.MustCompile(`https?:\/\/(?:[a-z\-]+\.)+[a-z]{2,6}(?:\/[^\/#?]+)+\.(?:jpe?g|gif|png)`)

	matches := imgRegex.FindStringSubmatch(i.Description)
	if len(matches) > 0 {
		i.ParsedImage = strings.Trim(matches[0], " ")
	}

	out <- 0
}

func ReutersParse(out chan<- int, i *ItemObject) {
	var tagRegex = regexp.MustCompile(`<div class=\"feedflare\">([^>]+)</div>`)
	i.Description = tagRegex.ReplaceAllString(i.Description, " ")
	removeAllTags(i)

	i.Description = strings.Trim(i.Description, "\n ")
	MetaParse(out, i)
}

func isSimilarButNotEqual(a, b string) bool {
	if a == b {
		return false
	}

	match := float64(0)
	for i := 0; i < len(b) && i < len(a); i++ {
		if b[i] == a[i] {
			match++
		} else {
			break
		}
	}
	count := float64(len(b))
	fmt.Printf("%f/%f=%f\n", match, count, match/count)
	return float64(match/count) > float64(0.75)
}

func MetaParse(out chan<- int, i *ItemObject) {
	resp, err := httpGet(1, i.Link)
	if err != nil {
		fmt.Printf("Connection error: %s\n", err)
		removeAllTags(i)
		out <- 0
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	bodyStr := string(body)

	html := new(Html)
	xml.Unmarshal([]byte(bodyStr), html)

	for _, meta := range html.Head.Meta {
		if meta.Property == "og:image" || meta.Name == "og:image" {
			i.ParsedImage = meta.Content
		}
	}

	if i.ParsedImage == "" {
		var imgRegex = regexp.MustCompile(`<[^>]*og:image[^>]*content=\"([^>]*)\"[^>]*>`)
		matches := imgRegex.FindStringSubmatch(bodyStr)
		if len(matches) > 1 {
			i.ParsedImage = matches[1]
		}
	}

	var imgRegex = regexp.MustCompile(`https?:\/\/(?:[a-z\-]+\.)+[a-z]{2,6}(?:\/[^\/#?]+)+\.(?:jpe?g|gif|png)`)
	matches := imgRegex.FindAllStringSubmatch(bodyStr, -1)
	for _, arr := range matches {
		for _, match := range arr {
			if isSimilarButNotEqual(match, i.ParsedImage) {
				i.ParsedImage = match
				fmt.Printf("Winner: %s\n", match)
			}
		}
	}

	if strings.Contains(i.ParsedImage, "template") ||
		strings.Contains(i.ParsedImage, "dnse-logo") ||
		strings.Contains(i.ParsedImage, "default.") ||
		strings.Contains(i.ParsedImage, "t_logo") ||
		strings.Contains(i.ParsedImage, "logo2login") {
		i.ParsedImage = ""
	}

	removeAllTags(i)

	out <- 0
}

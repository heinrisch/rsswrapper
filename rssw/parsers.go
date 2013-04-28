package rssw

import (
	"fmt"
	"github.com/opesun/goquery"
	"io/ioutil"
	"regexp"
	"strconv"
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

	MetaParse(out, i)
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

func getWidestImage(body string, i *ItemObject) {
	var imgRegex = regexp.MustCompile(`<img[^>]*src=\"([^>^"]*)\"[^>]*width=\"([0-9]*)\"[^>]*height=\"([0-9]*)\"[^>]*>`)
	matches := imgRegex.FindAllStringSubmatch(body, -1)
	minWidth := 150
	image := ""
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		if strings.Contains(match[1], "ad") || strings.Contains(match[1], "ybang") {
			continue
		}

		w, err := strconv.Atoi(match[2])
		if err != nil {
			fmt.Println(err)
			continue
		}

		h, err := strconv.Atoi(match[3])
		if err != nil {
			fmt.Println(err)
			continue
		}

		if w > minWidth && h > 100 {
			minWidth = w
			image = match[1]
		}

	}
	if image != "" {
		fmt.Printf("Changed from %s to %s\n", i.ParsedImage, image)
		i.ParsedImage = image
	}
}

func getOGImage(body string, i *ItemObject) {
	if i.ParsedImage == "" {
		var imgRegex = regexp.MustCompile(`<[^>]*og:image[^>]*content=\"([^>]*)\"[^>]*>`)
		matches := imgRegex.FindStringSubmatch(body)
		if len(matches) > 1 {
			i.ParsedImage = matches[1]
		}
	}
}

func MetaParse(out chan<- int, i *ItemObject) {
	resp, err := httpGet(2, i.Link)
	if err != nil {
		fmt.Printf("Connection error: %s\n", err)
		removeAllTags(i)
		out <- 0
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	bodyStr := string(body)

	getOGImage(bodyStr, i)
	//getWidestImage(bodyStr, i)

	nodes, err := goquery.Parse(bodyStr)

	if err != nil {
		fmt.Println(err)
	}

	nodes = nodes.Find("img")

	maxWidth := 200
	maxAlt := 100
	image := ""
	for _, node := range nodes {
		src, alt, width, _ := Attr(node)
		if !strings.HasPrefix(src, "http") {
			continue
		}

		if width > maxWidth {
			image = src
			maxWidth = width
		}

		if image == "" && len(alt) > maxAlt && width != 1 {
			image = src
			maxAlt = len(alt)
		}
	}

	if image != "" {
		fmt.Printf("Changed from %s to %s\n", i.ParsedImage, image)
		i.ParsedImage = image
	}

	if strings.Contains(i.ParsedImage, "template") ||
		strings.Contains(i.ParsedImage, "dnse-logo") ||
		strings.Contains(i.ParsedImage, "default.") ||
		strings.Contains(i.ParsedImage, "t_logo") ||
		strings.Contains(i.ParsedImage, "logo2login") ||
		strings.Contains(i.ParsedImage, "nprlogo") ||
		strings.Contains(i.ParsedImage, "ybang") {
		i.ParsedImage = ""
	}

	removeAllTags(i)

	out <- 0
}

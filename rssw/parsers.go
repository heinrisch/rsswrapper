package rssw

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/opesun/goquery"
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

func getFacebookStats(out chan<- int, i *ItemObject) {
	resp, err := httpGet(2, "http://api.facebook.com/restserver.php?method=links.getStats&urls="+i.Link)
	if err != nil {
		fmt.Printf("Connection error: %s\n", err)
		out <- 0
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	statResp := new(LinkStatResp)
	xml.Unmarshal(body, statResp)

	i.FacebookStats = statResp.Object

	out <- 0
}

func getTwitterStats(out chan<- int, i *ItemObject) {
	resp, err := httpGet(2, "http://urls.api.twitter.com/1/urls/count.json?url="+i.Link)
	if err != nil {
		fmt.Printf("Connection error: %s\n", err)
		out <- 0
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	count := new(CountObject)
	json.Unmarshal(body, count)

	i.TwitterStats = *count

	out <- 0
}

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

func getImageFromClass(body, class string, i *ItemObject) {
	nodes, err := goquery.Parse(body)

	if err != nil {
		fmt.Println(err)
		return
	}

	nodes = nodes.Find(class)
	getWidestImage(nodes.Html(), i)
}

func YahooParse(out chan<- int, i *ItemObject) {
	removeFirstImages(i)
	removeAllTags(i)

	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getMostSimilarAltImage(bodyStr, i)

	getImageFromClass(bodyStr, "div.yom-art-lead-img", i)

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

func SvdParse(out chan<- int, i *ItemObject) {
	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getMostSimilarAltImage(bodyStr, i)

	removeBadImage(i)

	removeAllTags(i)

	out <- 0
}

func NYTimesParse(out chan<- int, i *ItemObject) {
	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getMostSimilarAltImage(bodyStr, i)

	getImageFromClass(bodyStr, "div#article", i)

	removeBadImage(i)

	removeAllTags(i)

	out <- 0
}

func BBCParse(out chan<- int, i *ItemObject) {
	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getImageFromClass(bodyStr, "div.story-body", i)

	removeBadImage(i)

	removeAllTags(i)

	out <- 0
}
func ReutersParse(out chan<- int, i *ItemObject) {
	var tagRegex = regexp.MustCompile(`<div class=\"feedflare\">([^>]+)</div>`)
	i.Description = tagRegex.ReplaceAllString(i.Description, " ")
	removeAllTags(i)

	i.Description = strings.Trim(i.Description, "\n ")
	MetaParse(out, i)
}

func getOGImage(body string, i *ItemObject) {
	if i.ParsedImage == "" {
		var imgRegex = regexp.MustCompile(`<[^>]*og:image[^>]*content=\"([^>]*)\"[^>]*>`)
		matches := imgRegex.FindStringSubmatch(body)
		if len(matches) > 1 && isImageGood(matches[1]) {
			i.ParsedImage = matches[1]
		}
	}
}

func getPage(i *ItemObject) string {
	resp, err := httpGet(10, i.Link)
	if err != nil {
		fmt.Printf("Connection error: %s\n", err)
		return ""
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}

func GetSimilarityScore(a, b string) int {
	score := 0
	data := strings.Split(a, " ")
	for _, part := range data {
		if len(part) > 4 && strings.Contains(b, part) {
			score++
			index := strings.Index(b, part) + len(part)
			b = b[index:]
		}
	}
	return score
}

func getMostSimilarAltImage(bodyStr string, i *ItemObject) {
	nodes, err := goquery.Parse(bodyStr)

	if err != nil {
		fmt.Println(err)
		return
	}

	nodes = nodes.Find("img")

	maxScore := 2
	image := ""
	for _, node := range nodes {
		src, alt, width, _ := Attr(node)
		if !strings.HasPrefix(src, "http") {
			continue
		}

		score := GetSimilarityScore(i.Title, alt)
		if score > maxScore {
			image = src
			maxScore = width
		}
	}

	if image != "" {
		i.ParsedImage = image
	}
}

func getWidestImage(bodyStr string, i *ItemObject) {
	nodes, err := goquery.Parse(bodyStr)

	if err != nil {
		fmt.Println(err)
		return
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

		if width > maxWidth && isImageGood(src) {
			image = src
			maxWidth = width
		}

		if image == "" && len(alt) > maxAlt && width != 1 && isImageGood(src) {
			image = src
			maxAlt = len(alt)
		}
	}

	if image != "" {
		i.ParsedImage = image
	}
}

func MetaParse(out chan<- int, i *ItemObject) {
	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getMostSimilarAltImage(bodyStr, i)

	removeBadImage(i)

	removeAllTags(i)

	out <- 0
}

func removeBadImage(i *ItemObject) {
	if !isImageGood(i.ParsedImage) {
		i.ParsedImage = ""
	}
}

func isImageGood(img string) bool {
	badWords := [...]string{"template", "dnse-logo", "default.", "t_logo", "logo2login", "nprlogo", "ybang", "wasp", "ab66ddd94f78", "svdse_sidhuvud"}
	for _, word := range badWords {
		if strings.Contains(img, word) {
			return false
		}
	}
	return true
}

func getFacebookStats(out chan<- int, i *ItemObject) {
	resp, err := httpGet(5, "http://api.facebook.com/restserver.php?method=links.getStats&urls="+i.Link)
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
	resp, err := httpGet(5, "http://urls.api.twitter.com/1/urls/count.json?url="+i.Link)
	if err != nil {
		fmt.Printf("Connection error: %s\n", err)
		out <- 0
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	count := new(CountObject)
	json.Unmarshal(body, count)

	//Remove unreasonable high number, tiwtter calls has some problems
	if count.Count > 300000 {
		count.Count = 0
	}

	i.TwitterStats = *count

	out <- 0
}

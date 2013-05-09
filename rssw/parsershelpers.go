package rssw

import (
	"fmt"
	"github.com/opesun/goquery"
	"io/ioutil"
	"regexp"
	"strings"
)

func getImageFromClass(body, class string, i *ItemObject) {
	nodes, err := goquery.Parse(body)

	if err != nil {
		fmt.Println(err)
		return
	}

	nodes = nodes.Find(class)
	getWidestImage(nodes.Html(), i)
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

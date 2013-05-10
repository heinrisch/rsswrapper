package rssw

import (
	"encoding/json"
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

	bodyStr := getPage(i)

	getImageFromClass(bodyStr, "div.abArticle", i)

	out <- 0
}

func ExpressenParse(out chan<- int, i *ItemObject) {
	removeFirstImages(i)
	removeAllTags(i)

	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getImageFromClass(bodyStr, "div.b-article", i)

	out <- 0
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

func CNNParse(out chan<- int, i *ItemObject) {
	removeFirstImages(i)
	removeAllTags(i)

	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getImageFromClass(bodyStr, "div.cnn_strycntntlft", i)

	out <- 0
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

func MetaParse(out chan<- int, i *ItemObject) {
	bodyStr := getPage(i)

	getOGImage(bodyStr, i)

	getMostSimilarAltImage(bodyStr, i)

	removeBadImage(i)

	removeAllTags(i)

	out <- 0
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

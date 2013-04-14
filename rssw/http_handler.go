package rssw

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Start() {
	http.HandleFunc("/rss", rssHandler)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
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

	requests := r.URL.Query().Get("requests")
	if requests == "" {
		requests = "aftonbladet dn svd di svt reddit cnn bbc yahoo"
	}

	feeds := strings.Split(requests, " ")
	numberOfFeeds := len(feeds)

	channel := make(chan []ItemObject)
	for _, feed := range feeds {
		switch feed {
		case "aftonbladet":
			go getFeed(channel, "http://www.aftonbladet.se/rss.xml", AftonbladetParse)
			break
		case "dn":
			go getFeed(channel, "http://www.dn.se/nyheter/m/rss/", MetaParse)
			break
		case "svd":
			go getFeed(channel, "http://www.svd.se/?service=rss", MetaParse)
			break
		case "di":
			go getFeed(channel, "http://www.di.se/rss", MetaParse)
			break
		case "svt":
			go getFeed(channel, "http://www.svt.se/nyheter/regionalt/mittnytt/rss.xml", MetaParse)
			break
		case "reddit":
			go getFeed(channel, "http://www.reddit.com/r/gifs/.rss", RedditParse)
			break
		case "cnn":
			go getFeed(channel, "http://rss.cnn.com/rss/edition.rss", MetaParse)
			break
		case "bbc":
			go getFeed(channel, "http://feeds.bbci.co.uk/news/rss.xml", MetaParse)
			break
		case "yahoo":
			go getFeed(channel, "http://news.yahoo.com/rss/world", MetaParse)
			break
		case "reuters":
			go getFeed(channel, "http://feeds.reuters.com/reuters/topNews?format=xml", MetaParse)
			break
		default:
			numberOfFeeds--
		}
	}

	var items []ItemObject
	for i := 0; i < numberOfFeeds; i++ {
		rec := <-channel
		for _, i := range rec {
			if !strings.Contains(strings.ToUpper(i.Description), "NSFW") && !strings.Contains(strings.ToUpper(i.Title), "NSFW") {
				items = append(items, i)
			}
		}
	}

	sort.Sort(ByTime{items})

	res, _ := json.Marshal(items)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(res))
}

func getFeed(out chan<- []ItemObject, feed string, parser DescriptionParser) {
	resp, err := httpGet(5, feed)

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
			item.Source = rss.Channel.Link
			item.time = item.UnixTime()
			recentItems = append(recentItems, item)
		}
	}

	//Parse out description and image
	items := recentItems
	if parser != nil {
		parseChannel := make(chan int)
		toParse := 0
		for i := 0; i < len(items); i++ {
			toParse += 1
			go parser(parseChannel, &items[i])
		}

		for toParse > 0 {
			<-parseChannel
			toParse--
		}
	}

	out <- items

}

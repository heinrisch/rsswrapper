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

	channel := make(chan []ItemObject)

	feeds := 5
	go getFeed(channel, "http://www.aftonbladet.se/rss.xml", AftonbladetParse)
	go getFeed(channel, "http://www.dn.se/nyheter/m/rss/", nil)
	go getFeed(channel, "http://www.svd.se/?service=rss", SVDParse)
	go getFeed(channel, "http://www.reddit.com/r/gifs/.rss", RedditParse)
	go getFeed(channel, "http://rss.cnn.com/rss/edition.rss", nil)

	var items []ItemObject
	for i := 0; i < feeds; i++ {
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
			recentItems = append(recentItems, item)
		}
	}

	//Parse out description and image
	items := recentItems
	if parser != nil {
		for i := 0; i < len(items); i++ {
			parser(&items[i])
		}
	}

	out <- items

}

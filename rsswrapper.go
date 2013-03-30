package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Rss struct {
	All     string        `xml:",innerxml"`
	Channel ChannelObject `xml:"channel"`
}

type ChannelObject struct {
	All   string       `xml:",innerxml"`
	Items []ItemObject `xml:"item"`
}

type ItemObject struct {
	//All         string `xml:",innerxml"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Guid        string `xml:"guid"`
	Image       string `xml:"image"`
	Author      string `xml:"author"`
	Category    string `xml:"category"`
}

func main() {
	http.HandleFunc("/rss", rssHandler)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func rssHandler(w http.ResponseWriter, r *http.Request) {

	feeds := [...]string{
		"http://www.aftonbladet.se/rss.xml",
		"http://www.dn.se/nyheter/m/rss/",
		"http://www.svd.se/?service=rss",
		"http://www.reddit.com/r/gifs/.rss"}

	channel := make(chan []ItemObject)

	for _, feed := range feeds {
		fmt.Printf("getFeed %s\n", feed)
		go getFeed(channel, feed)
	}

	var items []ItemObject
	for _, _ = range feeds {
		rec := <-channel
		for _, i := range rec {
			items = append(items, i)
		}
	}

	res, _ := json.Marshal(items)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(res))
}

func getFeed(out chan<- []ItemObject, feed string) {
	resp, err := http.Get(feed)

	handleErr(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	rss := new(Rss)
	xml.Unmarshal(body, rss)

	out <- rss.Channel.Items

}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

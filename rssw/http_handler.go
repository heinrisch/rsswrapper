package rssw

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var DEFAULT_FEEDS string = "aftonbladet dn svd di svt reddit cnn bbc yahoo reuters nytimes npr expressen"

func Start() {
	http.HandleFunc("/rss", rssHandler)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+strconv.Itoa(4433), nil)
	if err != nil {
		panic(err)
	}
}

func createDatabaseIfNon(db *sql.DB) {
	//database not created
	_, err := db.Query("select * from news")

	if err != nil {
		fmt.Println("Creating database")
		_, err = db.Exec("create table news (id integer not null primary key, title unique, source text, time integer, trend integer, item text)")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getDatabase() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./news.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	return db
}

func WriteToDatabase() {
	timeDiff = 60 * 60 * 72

	db := getDatabase()

	createDatabaseIfNon(db)

	items := getItems(strings.Split(DEFAULT_FEEDS, " "))

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}

	stmt, err := tx.Prepare("insert into news(title, source, time, trend, item) values(?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return
	}
	updatestmt, err2 := tx.Prepare("update news set item=?, trend=? where title=? and source=?")
	if err2 != nil {
		fmt.Println(err2)
		return
	}

	defer stmt.Close()
	defer updatestmt.Close()
	for _, item := range items {
		jsonBlob, _ := json.Marshal(item)
		_, err = stmt.Exec(item.Title, item.Source, item.UnixTime(), item.TwitterStats.Count+item.FacebookStats.Shares, jsonBlob)
		if err != nil {
			fmt.Println(err)
			_, err = updatestmt.Exec(jsonBlob, item.TwitterStats.Count+item.FacebookStats.Shares, item.Title, item.Source)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	fmt.Println("Committing")
	tx.Commit()
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
		requests = DEFAULT_FEEDS
	}

	feeds := strings.Split(requests, " ")

	var items []ItemObject
	if r.URL.Query().Get("trending") == "" {
		items = getItemsFromDatabase(feeds)
	} else {
		items = getTrendingItemsFromDatabase(feeds)
	}

	res, _ := json.Marshal(items)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(res))
}

func getItemsFromQuery(query string) []ItemObject {
	db := getDatabase()
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	var items []ItemObject
	for rows.Next() {
		var jsonBlob []byte
		rows.Scan(&jsonBlob)
		var item ItemObject
		json.Unmarshal(jsonBlob, &item)
		items = append(items, item)
	}
	rows.Close()

	return items
}

func getTrendingItemsFromDatabase(feeds []string) []ItemObject {
	query := "select item from news where ("
	for i, feed := range feeds {
		query += "source='" + feed + "'"
		if i != len(feeds)-1 {
			query += " or "
		} else {
			query += ") "
		}
	}
	query += " and time > " + strconv.FormatInt(time.Now().Unix()-timeDiff, 10)
	query += " and trend > 5 and trend < 150000"
	query += " order by trend desc"
	query += " limit 50"

	return getItemsFromQuery(query)
}

func getItemsFromDatabase(feeds []string) []ItemObject {
	query := "select item from news where ("
	for i, feed := range feeds {
		query += "source='" + feed + "'"
		if i != len(feeds)-1 {
			query += " or "
		} else {
			query += ") "
		}
	}
	query += " and time > " + strconv.FormatInt(time.Now().Unix()-timeDiff, 10)
	query += " order by time desc"

	return getItemsFromQuery(query)
}

func getItems(feeds []string) []ItemObject {
	numberOfFeeds := len(feeds)

	channel := make(chan []ItemObject)
	for _, feed := range feeds {
		switch feed {
		case "aftonbladet":
			go getFeed(channel, feed, "http://www.aftonbladet.se/rss.xml", AftonbladetParse)
			break
		case "dn":
			go getFeed(channel, feed, "http://www.dn.se/nyheter/m/rss/", MetaParse)
			break
		case "svd":
			go getFeed(channel, feed, "http://www.svd.se/?service=rss", SvdParse)
			break
		case "di":
			go getFeed(channel, feed, "http://www.di.se/rss", MetaParse)
			break
		case "svt":
			go getFeed(channel, feed, "http://www.svt.se/nyheter/regionalt/mittnytt/rss.xml", MetaParse)
			break
		case "reddit":
			go getFeed(channel, feed, "http://www.reddit.com/r/gifs/.rss", RedditParse)
			break
		case "cnn":
			go getFeed(channel, feed, "http://rss.cnn.com/rss/edition.rss", CNNParse)
			break
		case "bbc":
			go getFeed(channel, feed, "http://feeds.bbci.co.uk/news/rss.xml", BBCParse)
			break
		case "yahoo":
			go getFeed(channel, feed, "http://news.yahoo.com/rss/world", YahooParse)
			break
		case "reuters":
			go getFeed(channel, feed, "http://feeds.reuters.com/reuters/topNews?format=xml", ReutersParse)
			break
		case "nytimes":
			go getFeed(channel, feed, "http://rss.nytimes.com/services/xml/rss/nyt/GlobalHome.xml", NYTimesParse)
			break
		case "npr":
			go getFeed(channel, feed, "http://www.npr.org/rss/rss.php?id=1001", MetaParse)
			break
		case "redditpics":
			go getFeed(channel, feed, "http://www.reddit.com/r/pics/.rss", RedditParse)
			break
		case "expressen":
			go getFeed(channel, feed, "http://www.expressen.se/Pages/OutboundFeedsPage.aspx?id=3642159&viewstyle=rss", ExpressenParse)
			break
		default:
			numberOfFeeds--
		}
	}

	var items []ItemObject
	statsChannel := make(chan int)
	for i := 0; i < numberOfFeeds; i++ {
		rec := <-channel
		for _, i := range rec {
			if !strings.Contains(strings.ToUpper(i.Description), "NSFW") && !strings.Contains(strings.ToUpper(i.Title), "NSFW") {
				items = append(items, i)
				go getFacebookStats(statsChannel, &items[len(items)-1]) // referecing i does not work since it is copied into array
				go getTwitterStats(statsChannel, &items[len(items)-1])
			}
		}
	}

	for i := 0; i < len(items); i++ {
		<-statsChannel
		<-statsChannel
	}

	return items
}

func getFeed(out chan<- []ItemObject, source, feed string, parser DescriptionParser) {
	resp, err := httpGet(15, feed)

	if err != nil {
		items := make([]ItemObject, 1)
		items[0].Title = err.Error()
		out <- items
		return
	}

	defer resp.Body.Close()
	d := xml.NewDecoder(resp.Body)
	d.CharsetReader = CharsetReader

	rss := new(Rss)
	err = d.Decode(rss)

	if err != nil {
		items := make([]ItemObject, 1)
		items[0].Title = err.Error()
		out <- items
		return
	}

	//Remove items older than 1 hour
	var recentItems []ItemObject
	for _, item := range rss.Channel.Items {
		if time.Now().Unix()-item.UnixTime() < timeDiff {
			item.Source = source
			item.Time = item.UnixTime()
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

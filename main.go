package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	_ "github.com/lib/pq"
)

func main() {
	// Wait for DB
	time.Sleep(2000 * time.Millisecond)

	var db *sql.DB

	dbFlags := flag.NewFlagSet("postgres", flag.ExitOnError)
	dbName := dbFlags.String("db", "", "Postgres Database")
	dbHost := dbFlags.String("host", "", "Postgres Host")
	dbPassword := dbFlags.String("password", "", "Postgres Password")
	dbUser := dbFlags.String("user", "", "Postgres User")

	dbFlags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(dbFlags, "POSTGRES")

	dbinfo := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
		*dbUser,
		*dbName,
		*dbPassword,
		*dbHost,
	)

	var err error
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal("Open failed:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Ping failed:", err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS urls (
			id serial NOT NULL,
			url text NOT NULL,
			CONSTRAINT urls_pkey PRIMARY KEY (id)
		) WITH (OIDS=FALSE);`)
	if err != nil {
		log.Fatal(err)
	}

	flags := flag.NewFlagSet("twitter", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")

	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Fatal("Consumer key and secret, and access token and secret required")
	}

	config := oauth1.NewConfig(*consumerKey, *consumerSecret)
	token := oauth1.NewToken(*accessToken, *accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	store := &URLStore{db}

	// make sure the client is working
	// see
	tweets, _, err := client.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
		Count: 20,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range tweets {
		log.Println(t.ID, t.User.ScreenName, t.Text)
		for _, url := range t.Entities.Urls {
			id, err := store.CreateURL(url.ExpandedURL)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("URL", id, t.ID, url.ExpandedURL)
		}
	}

	demux := twitter.NewSwitchDemux()
	demux.Tweet = TweetHandler(store)
	demux.Event = func(event *twitter.Event) {
		log.Println("EVENT ", event)
	}

	params := &twitter.StreamUserParams{
		With:          "followings",
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.User(params)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Ready")
	go demux.HandleChan(stream.Messages)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()
}

type DataStore interface {
	CreateURL(string) (int, error)
}

func TweetHandler(db DataStore) func(*twitter.Tweet) {
	return func(t *twitter.Tweet) {
		log.Println("called")
		recordType := "TWEET"

		if t.InReplyToStatusID > 0 {
			recordType = "REPLY"
		}

		if t.QuotedStatusID > 0 {
			recordType = "QUOTE"
		}

		if t.RetweetedStatus != nil {
			recordType = "RETWEET"
		}

		log.Println(recordType, t.ID, t.User.ScreenName, t.Text)
		for _, url := range t.Entities.Urls {
			id, err := db.CreateURL(url.ExpandedURL)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("URL", id, t.ID, url.ExpandedURL)
		}
	}
}

type URLStore struct {
	*sql.DB
}

func (db *URLStore) CreateURL(url string) (id int, err error) {
	err = db.QueryRow("INSERT INTO urls(url) VALUES($1) RETURNING id;", url).Scan(&id)
	return id, err
}

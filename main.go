package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	_ "github.com/lib/pq"
)

func main() {
	var db *sql.DB

	dbinfo := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
		"postgres",
		"postgres",
		"postgres",
		"postgres",
	)
	var err error
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
	demux := twitter.NewSwitchDemux()
	demux.Tweet = TweetHandler(store)

	params := &twitter.StreamUserParams{
		With:          "followings",
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.User(params)
	if err != nil {
		log.Fatal(err)
	}

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
		fmt.Println(t.Text)
		for _, url := range t.Entities.Urls {
			id, err := db.CreateURL(url.ExpandedURL)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("\t%d %s\n", id, url.ExpandedURL)
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

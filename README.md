# Twitter UserStream Client in Go

## Dependencies

- Go 1.5
- Docker Toolbox (tested with 1.10)
- Postgresql (latest/9.5)

## Run

Create a `.env` file under the project root with the following environment
variables and fill in the appropriate values:

```
export TWITTER_CONSUMER_KEY=""
export TWITTER_CONSUMER_SECRET=""
export TWITTER_ACCESS_TOKEN=""
export TWITTER_ACCESS_SECRET=""
```

Use `docker-compose` to build and run:

```
docker-compose build
docker-compose up
```

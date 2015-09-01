# Harvest Doc

Command line tool to generate a CSV file of Harvest fields and concepts.

## Install

```
go install github.com/chop.edu/harvestdoc
```

## HTTP

Run the server.
```
$ harvestdoc http
* Listening on :5000
```

POST a request with a JSON body containing the URL and optional token.

```
$ curl -X POST \
    -H "Accept: text/csv" \
    -H "Content-Type: application/json" \
    http://localhost:5000 -d '{
        "url": "http://harvest.research.chop.edu/demo/api/"
    }' > demo.csv
```

## CLI

```
$ harvestdoc http://harvest.research.chop.edu/demo/api/ > demo.csv
```

## Docker

Available on [Docker Hub](https://hub.docker.com/r/dbhi/harvestdoc/). The default command is to run the HTTP service on port 5000.

```
docker run -d --name harvestdoc -p 5000:5000 dbhi/harvestdoc
```

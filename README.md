# ChainDB

![Issues Tracker](https://img.shields.io/github/issues/Montessquio/ChainDB)
![MIT License](https://img.shields.io/github/license/Montessquio/ChainDB)

ChainDB is a searchable PDF Database featuring an attractive front-end, user accounts, and tag-based as well as full text search.

## Features

- [ ] Upload, tag, and search for documents all from one convenient web UI.
- [ ] Add user-defined tags to documents.
- [ ] Restrict which users can upload, edit tags, and delete records.
- [ ] Deploy Easily with Docker and Compose.

## Installation

ChainDB can be run as a standalone executable, or within a docker container.

## Deploy with Docker

**Docker compatibility is currently a work in progress. [See Tracking Issue (#2)](https://github.com/Montessquio/ChainDB/issues/2)**

Deploying with Docker is as simple as pulling and running the image.
You must already have an elasticsearch server running ([https://www.elastic.co/](https://www.elastic.co/)).

```bash
docker pull montessquio/ChainDB

docker run -d -p 80:80 \
-v HOST/DATA/PATH:/chain_data \
-e ES_URL="http://<elastic_url>:9200" \
--name chainDB montessquio/ChainDB
```

The following environment variables are available:

- `ES_URL` - The URL of the ElasticSearch database. Passed to the `-e` CLI Flag.
- `DOMAIN` - The domain that this utility is served under. Valid values include 192.168.2.1:80, app.example.net, and example.net. Passed to the `-r` CLI Flag.
- `ROOT` - The subdirectory of the domain this utility will be served under. If this is set to "db" then the page found normally at example.com/a.html will instead be found at example.com/db/a.html. Passed to the `-s` CLI Flag.

## Deploy with Docker-Compose

Deploying with Docker-Compose is easy. Download the compose file available at the root of the repo: [https://github.com/Montessquio/ChainDB/blob/master/docker-compose.yml](https://github.com/Montessquio/ChainDB/blob/master/docker-compose.yml)

Then, edit the file to suit your needs. All environment variables available in normal docker deployment are also available in docker-compose.

```bash
curl https://raw.githubusercontent.com/Montessquio/ChainDB/master/docker-compose.yml > docker-compose.yml

# Edit the docker file
vim docker-compose.yml

docker-compose up -d
```

ChainDB will be accessible via WebUI on port 80 by default.

## Deploy From Repo

### Compiling

Make sure you have the latest version of the Go toolchain installed: [https://golang.org/doc/install](https://golang.org/doc/install) as well as a functioning ElasticSearch instance: [https://www.elastic.co/](https://www.elastic.co/).

---

First, clone the repository to wherever you would like to run the server.

```plaintext
git clone https://github.com/Montessquio/ChainDB

cd ChainDB
```

Install all dependencies using `go get`

```plaintext
go get -v ./...
```

And then build the executable:

```plaintext
go build
```

The built executable should be found in the project root.

### Running

ChainDB needs two CLI arguments at minimum to run:

```plaintext
ChainDB -d <directory to store uploaded files> -e <ElasticSearch Server URL>
```

More configuration options can be found using `ChainDB -h`:

```plaintext
Usage of ChainDB:
  -d string
        Data directory to serve files from.
  -e string
        The URL of the ElasticSearch server.
  -p int
        Port to serve on (default 80)
  -r string
        The domain that this utility is served under. Valid values include 192.168.2.1:80, app.example.net, and example.net
  -s string
        The subdirectory of the domain this utility will be served under. If this is set to "db" then the page found normally at example.com/a.html will instead be found at example.com/db/a.html
```

ChainDB will be accessible via WebUI on port 80 by default.

FROM golang:1.14.6-alpine3.12
LABEL author="Nicolas \"Montessquio\" Suarez"

EXPOSE 80/tcp
EXPOSE 80/udp

RUN ["apk", "add", "git"]
WORKDIR "$GOPATH/src/github.com/Montessquio/ChainDB"
COPY backend/ webui/ www/ cli.go main.go go.mod go.sum "$GOPATH/src/github.com/Montessquio/ChainDB/"
ENV GO111MODULE=off
RUN go get -v github.com/spf13/afero && \
    go get -v github.com/sirupsen/logrus && \
    go get -v github.com/elastic/go-elasticsearch && \
    go build ./main.go

CMD "./ChainDB -e $ES_URL -r $DOMAIN -s $ROOT -p 80 -d /chain_data"

VOLUME /chain_data
FROM golang:1.14

COPY . /go/src/zabbix-matrix-bot/
RUN go get zabbix-matrix-bot/...
RUN go install zabbix-matrix-bot

CMD zabbix-matrix-bot

FROM golang:1.16.3
ENV GO111MODULE=on
WORKDIR /home/egc/FSRV_Edge/
COPY . .
RUN cd /home/egc/FSRV_Edge/ && go build -o edge
ENTRYPOINT ./edge
LABEL Name=SYSTEM Version=1.0
EXPOSE 8887

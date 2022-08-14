FROM golang:alpine AS build

RUN apk --no-cache add git ca-certificates

RUN git clone https://github.com/DowerX/live-server.git /live-server
WORKDIR /live-server
RUN go env -w GO111MODULE=off; \
    go get -u github.com/gorilla/mux; \
    go get -u github.com/lib/pq; \
    go get -u github.com/namsral/flag; \
    CGO_ENABLED=0 \
    go build \
    -installsuffix "static" \
    -o live-server

FROM scratch AS final
COPY --from=build /live-server/live-server /live-server
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/live-server"]
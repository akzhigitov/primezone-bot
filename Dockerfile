FROM golang:1.14.3 AS build
COPY . .
RUN go get -v -t -d
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /out/primezone .
FROM scratch AS bin
COPY --from=build /out/primezone /primezone
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/primezone"]
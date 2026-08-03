FROM docker.io/library/golang:1.26.5 AS builder
RUN apt update
RUN apt install nodejs npm ca-certificates -y
WORKDIR /app
COPY go.mod ./
COPY src/ ./src/
RUN pwd && ls -la && find /app -maxdepth 4
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ./fastlev ./src/cmd/fastlev/

FROM docker.io/library/alpine:3.23.5
COPY --from=builder /app/fastlev /usr/local/bin/fastlev
ENV FASTLEV_PATH=/usr/local/bin/fastlev
ENTRYPOINT ["fastlev"]
CMD ["--help"]
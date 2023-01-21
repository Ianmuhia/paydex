FROM golang:1.19-alpine AS builder
RUN apk --update --no-cache add git
RUN mkdir -p app
WORKDIR /app/

RUN apk add upx

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
# RUN CGO_ENABLED=0 go test -v ./...
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o -ldflags="-s -w" -o server .
RUN upx --best --lzma /app/server

FROM gcr.io/distroless/static:latest
WORKDIR /root
COPY --from=builder /app/server .

EXPOSE 8090
EXPOSE 8091
EXPOSE 8080

# Run the binary program produced by `go install`
ENTRYPOINT ["/root/server"]
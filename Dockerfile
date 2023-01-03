# Builder process as Kafka needs specific dependency to be built but not run
FROM golang:1.19-alpine AS builder

ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download
RUN apk -U add ca-certificates
RUN apk update && apk upgrade && apk add pkgconf git bash build-base sudo
RUN git clone https://github.com/edenhill/librdkafka.git && cd librdkafka && ./configure --prefix /usr && make && make install

COPY . .

RUN go build -tags musl --ldflags "-extldflags -static" -o user-service .

# Using a runner image
FROM alpine AS runner

COPY --from=builder /app /

EXPOSE 8080

CMD ["./user-service"]
FROM golang:1.19 as BUILDER

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY *.go ./
COPY internal/ ./internal

ENV CGO_ENABLED 0

RUN go build -o app .

FROM scratch

ENV HTTP_PORT 8080

EXPOSE 8080

COPY --from=BUILDER /app/app /

ENTRYPOINT [ "/app" ]

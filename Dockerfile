FROM golang:1.19-alpine as builder
RUN mkdir /go/src/app
WORKDIR /go/src/app
COPY . /go/src/app
RUN go mod download && go mod tidy && go build -o app .

FROM alpine:latest
COPY --from=builder /go/src/app/app .

CMD [ "./app" ]
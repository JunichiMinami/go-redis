# [go] dockerでredisの開発環境構築
まずgoの環境を作っていきます。
goの環境を初期化し、redisのモジュールをimportします。
```
go mod init go-redis
go get github.com/gomodule/redigo/redis
```

redisの動作を確認するためのgoのコードを書いていきます。
```
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
)

var pool *redis.Pool

func main() {
	addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	pool = &redis.Pool{
		MaxIdle: 30,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn := pool.Get()
		defer conn.Close()

		if _, err := conn.Do("SET", "key", "value"); err != nil {
			w.WriteHeader(500)
			return
		}
		value, err := redis.String(conn.Do("GET", "key"))
		if err != nil {
			w.WriteHeader(500)
			return
		}
		io.WriteString(w, value)
		w.WriteHeader(200)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

重要なところだけ抜粋
アドレスを指定し、redisを初期化します。
```
addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
pool = &redis.Pool{
  MaxIdle: 30,
  Dial: func() (redis.Conn, error) {
    return redis.Dial("tcp", addr)
  },
}
```

httpハンドラを作成します。
redisに接続し、キーがkey,値がvalueの文字列を設定します。
すぐにそのキーを使って値を取り出し、レスポンスします。
```
  http.HandleFunc("/redis", func(w http.ResponseWriter, r *http.Request) {
		conn := pool.Get()
		defer conn.Close()

		if _, err := conn.Do("SET", "key", "value"); err != nil {
			w.WriteHeader(500)
			return
		}
		value, err := redis.String(conn.Do("GET", "key"))
		if err != nil {
			w.WriteHeader(500)
			return
		}
		io.WriteString(w, value)
		w.WriteHeader(200)
	})
```

次にDockerfileを作成します。
goの環境を作成する
```
FROM golang:1.19-alpine as builder
RUN mkdir /go/src/app
WORKDIR /go/src/app
COPY . /go/src/app
RUN go mod download && go mod tidy && go build -o app .

FROM alpine:latest
COPY --from=builder /go/src/app/app .

CMD [ "./app" ]
```

次にgoの環境とredisを接続するための
docker-compose.ymlファイルを作成します。
```
version: '3.4'

services:
  app:
    container_name: app
    build: .
    ports:
      - "8080:8080"
    environment:
      REDIS_HOST: redis
      REDIS_PORT: 6379
    depends_on:
      - redis
  
  redis:
    container_name: redis
    image: "redis:7.0"
    ports:
      - "6379:6379"
    volumes:
      - "./redis/db:/data"
```
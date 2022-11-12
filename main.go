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

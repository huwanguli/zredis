package main

import (
	"log"

	"zredis/core"
)

func main() {
	srv := core.NewServer()
	log.Fatal(srv.Listen(":6379"))
}

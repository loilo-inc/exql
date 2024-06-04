package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

func main() {
	arch := runtime.GOARCH
	var prefix string
	switch arch {
	case "amd64":
		prefix = "amd64"
	case "arm64":
		prefix = "arm64v8"
	}
	if prefix == "" {
		log.Fatalf("unsupported arch: %s", arch)
	}
	yml := fmt.Sprintf(`
version: "3.7"
services:
  mysql:
    container_name: exql_mysql8
    image: %s/mysql:8
    ports:
      - 13326:3306
    environment:
      MYSQL_ALLOW_EMPTY_PASSWORD: 1
      MYSQL_DATABASE: exql
    volumes:
      - ./schema:/docker-entrypoint-initdb.d`, prefix)
	err := os.WriteFile("compose.yml", []byte(yml), 0644)
	if err != nil {
		log.Fatalf("failed to write compose.yml: %s", err)
	}
}

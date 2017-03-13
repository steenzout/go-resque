package main

import (
	"os"
	"strconv"

	"github.com/steenzout/go-resque/test/log"
)

const (
	// Package package name.
	Package = "test.integration.main"

	// RedisServer name of the environment variable that holds the name of the Redis server.
	RedisServer = "RedisServer"
	// RedisPort name of the environment variable that holds the port under which Redis is running.
	RedisPort = "RedisPort"
)

var (
	// server is the name of the Redis server.
	server string
	// port is the port under which Redis is running.
	port int
)

func init() {
	initRedisServer()
	initRedisPort()
}

func initRedisPort() {
	portEnv := os.Getenv(RedisPort)
	log.Infof(Package, "env %s=%s", RedisPort, portEnv)

	if portEnv == "" {
		port = 6379
	} else {
		v, err := strconv.Atoi(portEnv)
		if err != nil {
			port = 6379
		} else {
			port = v
		}
	}

	log.Infof(Package, "port=%d", port)
}

func initRedisServer() {
	server = os.Getenv(RedisServer)
	log.Infof(Package, "env %s=%s", RedisServer, server)

	if server == "" {
		server = "127.0.0.1"
	}
	log.Infof(Package, "server=%s", server)
}

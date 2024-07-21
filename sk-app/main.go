package main

import (
	"micro/pkg/bootstrap"
	"micro/sk-app/setup"
)

func main() {
	setup.InitZk()
	setup.InitRedis()
	setup.InitServer(bootstrap.HttpConfig.Host, bootstrap.HttpConfig.Port)

}

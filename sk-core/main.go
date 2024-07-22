package main

import "micro/sk-core/setup"

func main() {

	setup.InitZk()
	setup.InitRedis()
	setup.RunService()

}

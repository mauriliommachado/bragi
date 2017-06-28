package main

import (
	"./server"
	"./controllers"
)

func main() {
	controllers.Run()
	server.Start(server.ServerProperties{Address: "/bragi", Port: "8083"})
}

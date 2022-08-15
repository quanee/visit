package main

type state struct {
	status State
}

type node struct {
	state
	id    uint64
	nodes map[string]*node
	ip    string
	port  string
}

func Ping()    {}
func Pong()    {}
func Publish() {}

func main() {

}

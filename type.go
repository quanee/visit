package main

type State uint8

const (
	_ State = iota
	FAIL
	ONLINE
)

type MsgType uint8

const (
	_ MsgType = iota
	PING
	PONG
	PUBLISH
	MEET
)

type Node struct {
	status    State
	id        uint64
	fadeout   uint64
	count     uint64
	interval  uint64 // s
	syncIndex uint64
	nodes     map[string]*Node
	ip        string
	port      string
}

type Msg struct {
	payload []byte
}

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

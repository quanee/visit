package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"time"
)

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
	MEET
	MOOT
	SYNC
	PUBLISH
)

type Msg struct {
	MsgType
	AddrPort netip.AddrPort
	Payload  []byte `json:"payload"`
}

type MeetMSG struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type MootMSG struct {
	ID string `json:"id"`
}

type Node struct {
	status    State
	ID        string `json:"id"`
	fadeout   uint64
	count     uint64
	interval  uint64 // s
	syncIndex uint64
	nodes     map[string]*Node
	ip        string
	port      uint16
	ctx       context.Context
	mux       sync.RWMutex
	conn      *net.UDPConn
	msgChan   chan *Msg
}

func (n *Node) tickMsg() {
	ticker := time.NewTicker(time.Duration(n.interval) * time.Second)
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			for _, node := range n.nodes {
				n.Ping(node)
			}
		}
	}
}

func (n *Node) ReadMsg() {
	go n.readMsg()
}

func (n *Node) readMsg() {
	for {
		body := make([]byte, 4096)
		length, raddr, err := n.conn.ReadFromUDPAddrPort(body)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(raddr)
		msg := &Msg{}
		if err := json.Unmarshal(body[:length], msg); err != nil {
			log.Println(err)
			continue
		}
		msg.AddrPort = raddr
		n.msgChan <- msg
	}
}

func (n *Node) ProcessMsg() {
	go n.processMsg()
}

func (n *Node) processMsg() {
	for {
		select {
		case msg := <-n.msgChan:
			switch msg.MsgType {
			case MEET:
				log.Printf("read meet msg: %v:%v\n", msg.AddrPort.Addr().String(), msg.AddrPort.Port())
				if err := n.addNode(msg); err != nil {
					log.Println(err)
					continue
				}
			case MOOT:
				moot := MootMSG{}
				if err := json.Unmarshal(msg.Payload, &moot); err != nil {
					log.Println(err)
					continue
				}
				n.ID = moot.ID
				log.Println("add successful, id: ", n.ID)
			case PING:
				log.Printf("%v: %v\n", msg.AddrPort, string(msg.Payload))
				n.Pong(msg.AddrPort)
			case PONG:
				log.Printf("%v: %v\n", msg.AddrPort, string(msg.Payload))
			case PUBLISH:
			case SYNC:
			}
		}
	}
}

func (n *Node) addNode(msg *Msg) error {
	id := uuid()
	n.nodes[id] = &Node{
		ip:   msg.AddrPort.Addr().String(),
		port: msg.AddrPort.Port(),
	}
	moot := MootMSG{}
	moot.ID = id
	body, err := json.Marshal(moot)
	if err != nil {
		return err
	}
	rmsg := Msg{
		MsgType: MOOT,
		Payload: body,
	}
	response, err := json.Marshal(rmsg)
	if err != nil {
		return err
	}
	if _, err := n.conn.WriteToUDPAddrPort(response, msg.AddrPort); err != nil {
		return err
	}

	log.Printf("add node successful, id: %v", moot.ID)
	return nil
}

func (n *Node) Meet(mip, mport string) error {
	msg := Msg{MsgType: MEET}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	udpaddr, err := net.ResolveUDPAddr("udp", mip+":"+mport)
	if err != nil {
		return err
	}

	if _, err := n.conn.WriteToUDP(body, udpaddr); err != nil {
		return err
	}
	return nil
}

func (n *Node) Ping(node *Node) {
	port := strconv.Itoa(int(node.port))
	addr, err := net.ResolveUDPAddr("udp", node.ip+":"+port)
	if err != nil {
		log.Println(err)
		return
	}
	ping := Msg{MsgType: PING, AddrPort: addr.AddrPort(), Payload: []byte("PING")}
	body, err := json.Marshal(ping)
	if err != nil {
		log.Println(err)
		return
	}
	if _, err := n.conn.WriteToUDP(body, addr); err != nil {
		log.Println(err)
		return
	}
}

func (n *Node) Pong(port netip.AddrPort) {
	pong := Msg{
		MsgType: PONG,
		Payload: []byte("PONG"),
	}
	body, err := json.Marshal(pong)
	if err != nil {
		log.Println(err)
		return
	}
	if _, err := n.conn.WriteToUDPAddrPort(body, port); err != nil {
		log.Println(err)
		return
	}
}

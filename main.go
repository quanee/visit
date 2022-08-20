package main

import (
	"context"
	"flag"
	"log"
	"net"
)

func (n *Node) Ping()    {}
func (n *Node) Pong()    {}
func (n *Node) Publish() {}
func (n *Node) TickMsg() {
	go n.tickMsg()
}

func main() {
	n := &Node{
		ctx:      context.Background(),
		nodes:    make(map[string]*Node),
		msgChan:  make(chan *Msg, 10000),
		interval: 5,
	}
	var mip, mport, lip, lport string
	var meeted bool
	flag.StringVar(&mip, "mip", "127.0.0.1", "current node first meet cluster node's ip")
	flag.StringVar(&mport, "mport", "63999", "current node first meet cluster node's port")
	flag.StringVar(&lip, "lip", "127.0.0.1", "current node local cluster node's ip")
	flag.StringVar(&lport, "lport", "20493", "current node local cluster node's port")
	flag.BoolVar(&meeted, "meeted", false, "current node is first node of cluster")
	flag.Uint64Var(&n.interval, "interval", 5, "node sync interval second")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	updaddr, err := net.ResolveUDPAddr("udp", lip+":"+lport)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", updaddr)
	if err != nil {
		log.Fatal(err)
	}
	n.conn = conn
	log.Printf("listing at %v:%v", lip, lport)
	if meeted {
		if err := n.Meet(mip, mport); err != nil {
			log.Fatal(err)
		}
	}

	n.TickMsg()
	n.ReadMsg()
	n.ProcessMsg()
	for {
	}
}

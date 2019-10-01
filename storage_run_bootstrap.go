package main 

import (
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/server"
	"Go-DAG-storageNode/query"
	"Go-DAG-storageNode/Discovery"
	"Go-DAG-storageNode/storage"
	"net"
	"fmt"
	"time"
	"math/rand"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var peers dt.Peers
	var genesis dt.Transaction
	var signatur []byte
	copy(genesis.Txid[:],[]byte("1234567812345678"))
	storage.AddTransaction(genesis,signatur)
	peers.Fds = make(map[string] net.Conn)
	var srv server.Server
	srv.Peers = &peers
	go srv.StartServer()
	time.Sleep(time.Second)
	ips := Discovery.GetIps("169.254.175.29:8000")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	fmt.Println("started storage node")
	fmt.Println("=============================")
	query.StartServer()
}
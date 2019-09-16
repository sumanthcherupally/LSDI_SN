package main 

import (
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/server"
	"Go-DAG-storageNode/query"
	"Go-DAG-storageNode/Discovery"
	"Go-DAG-storageNode/sync"
	//"Go-DAG-storageNode/storage"
	"net"
	"fmt"
	"time"
	"math/rand"
	"strings"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var peers dt.Peers
	peers.Fds = make(map[string] net.Conn)
	var srv server.Server
	srv.Peers = &peers
	go srv.StartServer()
	time.Sleep(time.Second)
	ips := Discovery.GetIps("192.168.122.190:8000")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	fmt.Println("connection established with all peers")
	time.Sleep(time.Second)
	hashes := sync.RequestHashes(peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]])
	fmt.Println(len(hashes))
	sync.QueryTransactions(peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]],hashes)
	// for _,node := range storage.OrphanedTransactions {
	// 	storage.AddTransaction(node.Tx,node.Signature)
	// }
	fmt.Println("Database synced")
	query.StartServer()
}

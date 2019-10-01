package main 

import (
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/server"
	"Go-DAG-storageNode/query"
	"Go-DAG-storageNode/Discovery"
	"Go-DAG-storageNode/sync"
	"Go-DAG-storageNode/storage"
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
	fmt.Println("=============================\n")
	fmt.Println("discovering other nodes..")
	time.Sleep(3*time.Second)
	ips := Discovery.GetIps("169.254.175.29:8000")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	fmt.Println("connected to peers")
	fmt.Println("=============================\n")
	fmt.Println("syncing database..")
	time.Sleep(2*time.Second)
	hashes := sync.RequestHashes(peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]])
	fmt.Println(len(hashes))
	sync.QueryTransactions(peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]],hashes)
	// for _,node := range storage.OrphanedTransactions {
	// 	storage.AddTransaction(node.Tx,node.Signature)
	// }
	fmt.Println(len(storage.OrphanedTransactions))
	fmt.Println("database updated")
	time.Sleep(2*time.Second)
	fmt.Println("storage node started")
	fmt.Println("=============================\n")
	query.StartServer()
}

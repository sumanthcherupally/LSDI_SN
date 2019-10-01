package main 

import (
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/server"
	"Go-DAG-storageNode/query"
	"Go-DAG-storageNode/Discovery"
	"Go-DAG-storageNode/sync"
	// "Go-DAG-storageNode/storage"
	log "Go-DAG-storageNode/logdump"
	"net"
	// "log"
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
	log.Println("discovering other nodes..")
	time.Sleep(3*time.Second)
	ips := Discovery.GetIps("169.254.175.29:8000")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	log.Println("connected to peers")
	log.Println("syncing database..")
	time.Sleep(2*time.Second)
	hashes := sync.RequestHashes(peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]])
	// log.Println(len(hashes))
	sync.QueryTransactions(peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]],hashes)
	// for _,node := range storage.OrphanedTransactions {
	// 	storage.AddTransaction(node.Tx,node.Signature)
	// }
	// log.Println(len(storage.OrphanedTransactions))
	log.Println("database updated")
	time.Sleep(2*time.Second)
	log.Println("storage node started")
	query.StartServer()
}

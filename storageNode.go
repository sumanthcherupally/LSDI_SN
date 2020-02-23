package main

import (
	"Go-DAG-storageNode/Crypto"
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/database"
	"Go-DAG-storageNode/node"
	"Go-DAG-storageNode/p2p"
	"Go-DAG-storageNode/query"
	"Go-DAG-storageNode/storage"
	"os"
)

func main() {
	var PrivateKey Crypto.PrivateKey
	if Crypto.CheckForKeys() {
		PrivateKey = Crypto.LoadKeys()
	} else {
		PrivateKey = Crypto.GenerateKeys()
	}
	database.OpenDB()
	var ID p2p.PeerID
	ID.PublicKey = Crypto.SerializePublicKey(&PrivateKey.PublicKey)
	v := constructGenisis()
	storage.AddTransaction(v.Tx, v.Signature)
	var ch chan p2p.Msg
	if os.Args[1] == "b" {
		ch = node.NewBootstrap(ID)
	} else if os.Args[1] == "n" {
		ch = node.New(ID)
	}
	query.StartServer()
}

func constructGenisis() dt.Vertex {
	var tx dt.Transaction
	tx.Hash = Crypto.Hash([]byte("IOT BLOCKCHAIN GENISIS"))
	var v dt.Vertex
	v.Tx = tx
	v.Signature = make([]byte, 72)
	return v
}

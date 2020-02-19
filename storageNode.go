package main

import (
	"DAG-SN/Crypto"
	dt "DAG-SN/DataTypes"
	"DAG-SN/database"
	"DAG-SN/node"
	"DAG-SN/p2p"
	"DAG-SN/query"
	"DAG-SN/storage"
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
	if os.Args[1] == "b" {
		node.NewBootstrap(ID)
	} else if os.Args[1] == "n" {
		node.New(ID)
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

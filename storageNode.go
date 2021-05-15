package main

import (
	"LSDI_SN/Crypto"
	dt "LSDI_SN/DataTypes"
	"LSDI_SN/database"
	"LSDI_SN/node"
	"LSDI_SN/p2p"
	"LSDI_SN/query"
	"LSDI_SN/storage"
)

func main() {
	var PrivateKey Crypto.PrivateKey
	if Crypto.CheckForKeys() {
		PrivateKey = Crypto.LoadKeys()
	} else {
		PrivateKey = Crypto.GenerateKeys()
	}
	var ID p2p.PeerID
	ID.PublicKey = Crypto.SerializePublicKey(&PrivateKey.PublicKey)
	v := constructGenisis()
	db := database.OpenDB()
	defer database.CloseDB(db)
	storageCh := make(chan dt.ForwardTx, 20)
	ch := node.New(&ID, storageCh, db)
	// initializing the storage layer
	var st storage.Server
	st.ForwardingCh = ch
	st.ServerCh = storageCh
	st.DB = db
	st.AddTransaction(v.Tx, v.Signature)
	go st.Run()
	query.Run(db)
}

func constructGenisis() dt.Vertex {
	var tx dt.Transaction
	tx.Hash = Crypto.Hash([]byte("IOT BLOCKCHAIN GENISIS"))
	var v dt.Vertex
	v.Tx = tx
	v.Signature = make([]byte, 72)
	return v
}

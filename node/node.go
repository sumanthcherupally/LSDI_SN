package node

import (
	dt "LSDI_SN/DataTypes"
	"LSDI_SN/p2p"
	"LSDI_SN/serialize"
	"LSDI_SN/storage"
	"log"
)

// New is used to create the blockchain node
// It spins up a server to accept transactions and returns a channel for client to broadcast transactions
// Input : p2pID(contains IP address), PrivKey(PrivateKey)
// Output : channel that accepts transaction to be broadcasted
func New(hostID *p2p.PeerID, sch chan dt.ForwardTx, db storage.DB) chan p2p.Msg {

	var srv p2p.Server
	srv.HostID.PublicKey = hostID.PublicKey
	hostID = &srv.HostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	srv.ShardingSignal = make(chan dt.ShardSignal)
	srv.ShardTransactions = make(chan dt.ShardTransaction)
	go srv.Run()

	go func() {
		for {
			p := <-srv.NewPeer
			go handle(&p, srv.BroadcastMsg, srv.ShardingSignal, srv.ShardTransactions, srv.RemovePeer, sch, db)
		}
	}()
	return srv.BroadcastMsg
}

func handleMsg(msg p2p.Msg, send chan p2p.Msg, p *p2p.Peer, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction, sch chan dt.ForwardTx, db storage.DB) {
	// check for transactions or request for transactions
	if msg.ID == 32 {
		// transaction
		tx, sign := serialize.Decode32(msg.Payload, msg.LenPayload)
		if validTransaction(tx, sign) {
			var sTx dt.ForwardTx
			sTx.Tx = tx
			sTx.Signature = sign
			sTx.Peer = p.GetPeerConn()
			sTx.Forward = true
			sch <- sTx

		}
	} else if msg.ID == 34 {
		hash := msg.Payload
		var replyMsg p2p.Msg
		replyMsg.ID = 33
		if storage.CheckifPresentDb(db, hash) {
			tx, sign := storage.GetTransactiondb(db, hash)
			replyMsg.Payload = append(serialize.Encode(tx), sign...)
			replyMsg.LenPayload = uint32(len(replyMsg.Payload))
			p.Send(replyMsg)
		}
	} else if msg.ID == 33 {
		// reply to a request sync transaction
		tx, sign := serialize.Decode32(msg.Payload, msg.LenPayload)
		if validTransaction(tx, sign) {
			var sTx dt.ForwardTx
			sTx.Tx = tx
			sTx.Signature = sign
			sTx.Peer = p.GetPeerConn()
			sTx.Forward = false
			sch <- sTx
		}
	} else if msg.ID == 36 {
		tx, _ := serialize.Decode36(msg.Payload, msg.LenPayload)
		Shardtxch <- tx
	} else if msg.ID == 35 {
		signal, _ := serialize.Decode35(msg.Payload, msg.LenPayload)
		select {
		case ShardSignalch <- signal:
		default:
		}
	}
}

// read the messages and handle
func handle(p *p2p.Peer, send chan p2p.Msg, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction, errChan chan p2p.Peer, sch chan dt.ForwardTx, db storage.DB) {
	for {
		msg, err := p.GetMsg()
		if err != nil {
			errChan <- *p
			log.Println(err)
			break
		}
		handleMsg(msg, send, p, ShardSignalch, Shardtxch, sch, db)
	}
}

func validTransaction(tx dt.Transaction, sign []byte) bool {
	return true
}

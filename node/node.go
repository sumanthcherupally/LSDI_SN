package node

import (
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/p2p"
	"Go-DAG-storageNode/serialize"
	"Go-DAG-storageNode/storage"
	"log"
)

// New ...
func New(hostID *p2p.PeerID, sch chan dt.ForwardTx) chan p2p.Msg {

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
			go handle(&p, srv.BroadcastMsg, srv.ShardingSignal, srv.ShardTransactions, srv.RemovePeer, sch)
		}
	}()
	return srv.BroadcastMsg
}

func handleMsg(msg p2p.Msg, send chan p2p.Msg, p *p2p.Peer, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction, sch chan dt.ForwardTx) {
	// check for transactions or request for transactions
	if msg.ID == 32 {
		// transaction
		tx, sign := serialize.Decode32(msg.Payload, msg.LenPayload)
		if validTransaction(tx, sign) {
			// tr := storage.AddTransaction(tx, sign)
			// if tr == 1 {
			// 	send <- msg
			// logLock.Lock()
			// f.WriteString(fmt.Sprintf("%d %d %d\n", p.ID.IP, time.Now().Minute(), time.Now().Second()))
			// logLock.Unlock()
			// 	// fmt.Println(p.ID.IP)
			// 	// fmt.Println(time.Now())
			// } else if tr == 2 {
			// var msg p2p.Msg
			// msg.ID = 34
			// if !storage.CheckifPresentDb(tx.LeftTip[:]) {
			// 	msg.Payload = tx.LeftTip[:]
			// 	msg.LenPayload = uint32(len(msg.Payload))
			// 	p.Send(msg)
			// }
			// if !storage.CheckifPresentDb(tx.RightTip[:]) {
			// 	msg.Payload = tx.RightTip[:]
			// 	msg.LenPayload = uint32(len(msg.Payload))
			// 	p.Send(msg)
			// }
			// } else if tr == 0 {
			// 	fmt.Println("Duplicate detected")
			// }
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
		if storage.CheckifPresentDb(hash) {
			tx, sign := storage.GetTransactiondb(hash)
			replyMsg.Payload = append(serialize.Encode(tx), sign...)
		}
		replyMsg.LenPayload = uint32(len(replyMsg.Payload))
		p.Send(replyMsg)
	} else if msg.ID == 33 {
		// reply to a request sync transaction
		tx, sign := serialize.Decode32(msg.Payload, msg.LenPayload)
		if validTransaction(tx, sign) {
			// tr := storage.AddTransaction(tx, sign)
			// if tr == 2 {
			// 	var msg p2p.Msg
			// 	msg.ID = 34
			// 	if !storage.CheckifPresentDb(tx.LeftTip[:]) {
			// 		msg.Payload = tx.LeftTip[:]
			// 		msg.LenPayload = uint32(len(msg.Payload))
			// 		p.Send(msg)
			// 	}
			// 	if !storage.CheckifPresentDb(tx.RightTip[:]) {
			// 		msg.Payload = tx.RightTip[:]
			// 		msg.LenPayload = uint32(len(msg.Payload))
			// 		p.Send(msg)
			// 	}
			// }
			var sTx dt.ForwardTx
			sTx.Tx = tx
			sTx.Signature = sign
			sTx.Peer = p.GetPeerConn()
			sTx.Forward = false
			sch <- sTx
		}
	} else if msg.ID == 36 {
		tx, _ := serialize.Decode36(msg.Payload, msg.LenPayload)
		// if sh.VerifyShardTransaction(tx, sign, 4) {
		// 	Shardtxch <- tx
		// }
		Shardtxch <- tx
	}
}

// read the messages and handle
func handle(p *p2p.Peer, send chan p2p.Msg, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction, errChan chan p2p.Peer, sch chan dt.ForwardTx) {
	for {
		msg, err := p.GetMsg()
		if err != nil {
			errChan <- *p
			log.Println(err)
			break
		}
		handleMsg(msg, send, p, ShardSignalch, Shardtxch, sch)
	}
}

func validTransaction(tx dt.Transaction, sign []byte) bool {
	return true
}

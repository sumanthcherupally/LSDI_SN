package node

import (
	dt "DAG-SN/DataTypes"
	"DAG-SN/p2p"
	"DAG-SN/serialize"
	"DAG-SN/storage"
	"GO-DAG/Crypto"
	"log"
)

// New ...
func New(hostID p2p.PeerID) {

	var srv p2p.Server
	srv.HostID = hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	go srv.Run()

	go func() {
		for {
			p := <-srv.NewPeer
			go handle(p, srv.BroadcastMsg, srv.RemovePeer)
		}
	}()
	// sync.Sync(srv.GetRandomPeer())
	return
}

// NewBootstrap ...
func NewBootstrap(hostID p2p.PeerID) {
	log.Println("BOOTSTRAP NODE")
	var srv p2p.Server
	srv.HostID = hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	go srv.Run()
	go func() {
		for {
			p := <-srv.NewPeer
			go handle(p, srv.BroadcastMsg, srv.RemovePeer)
		}
	}()
	return
}

func handleMsg(msg p2p.Msg, send chan p2p.Msg, p p2p.Peer) {
	// check for transactions or request for transactions
	if msg.ID == 32 {
		// transaction
		tx, sign := serialize.DeserializeTransaction(msg.Payload, msg.LenPayload)
		log.Println(Crypto.EncodeToHex(tx.Hash[:]))
		if validTransaction(tx, sign) {
			tr := storage.AddTransaction(tx, sign)
			if tr == 1 {
				send <- msg
			} else if tr == 2 {
				var msg p2p.Msg
				msg.ID = 34
				if !storage.CheckifPresentDb(tx.LeftTip[:]) {
					msg.Payload = tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
				if !storage.CheckifPresentDb(tx.RightTip[:]) {
					msg.Payload = tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
			}
		}
	} else if msg.ID == 34 {
		hash := msg.Payload
		var replyMsg p2p.Msg
		replyMsg.ID = 32
		if storage.CheckifPresentDb(hash) {
			tx, sign := storage.GetTransactiondb(hash)
			replyMsg.Payload = append(serialize.Encode(tx), sign...)
		}
		replyMsg.LenPayload = uint32(len(replyMsg.Payload))
		p.Send(replyMsg)
	}
}

// read the messages and handle
func handle(p p2p.Peer, send chan p2p.Msg, errChan chan p2p.Peer) {
	for {
		msg, err := p.GetMsg()
		if err != nil {
			errChan <- p
			log.Println(err)
			break
		}
		go handleMsg(msg, send, p)
	}
}

func validTransaction(tx dt.Transaction, sign []byte) bool {
	return true
}

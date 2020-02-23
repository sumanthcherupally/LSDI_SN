package node

import (
	"Go-DAG-storageNode/Crypto"
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/p2p"
	"Go-DAG-storageNode/serialize"
	"Go-DAG-storageNode/storage"
	"log"
)

// New ...
func New(hostID *p2p.PeerID) chan p2p.Msg {

	var srv p2p.Server
	srv.HostID.PublicKey = hostID.PublicKey
	hostID = &srv.HostID
	// srv.HostID = hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	srv.ShardTransactions = make(chan dt.ShardTransaction)
	srv.ShardingSignal = make(chan dt.ShardSignal)
	go srv.Run()

	go func() {
		for {
			p := <-srv.NewPeer
			go handle(&p, srv.BroadcastMsg, srv.ShardingSignal, srv.ShardTransactions, srv.RemovePeer)
			// go handle(p, srv.BroadcastMsg, srv.RemovePeer)
		}
	}()
	// sync.Sync(srv.GetRandomPeer())
	return srv.BroadcastMsg
}

// NewBootstrap ...
func NewBootstrap(hostID *p2p.PeerID) chan p2p.Msg {
	log.Println("BOOTSTRAP NODE")
	var srv p2p.Server
	srv.HostID.PublicKey = hostID.PublicKey
	hostID = &srv.HostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	srv.ShardTransactions = make(chan dt.ShardTransaction)
	srv.ShardingSignal = make(chan dt.ShardSignal)
	go srv.Run()
	go func() {
		for {
			p := <-srv.NewPeer
			go handle(&p, srv.BroadcastMsg, srv.ShardingSignal, srv.ShardTransactions, srv.RemovePeer)
		}
	}()
	return srv.BroadcastMsg
}

func handleMsg(msg p2p.Msg, send chan p2p.Msg, p *p2p.Peer, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction) {
	// check for transactions or request for transactions
	if msg.ID == 32 {
		// transaction
		tx, sign := serialize.Decode32(msg.Payload, msg.LenPayload)
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
	} else if msg.ID == 35 { //Shard signal
		signal, _ := serialize.Decode35(msg.Payload, msg.LenPayload)
		select {
		case ShardSignalch <- signal:
			send <- msg
		default:
			log.Println("Duplicate sharding signal")
		}
	} else if msg.ID == 36 { //Sharding tx from other nodes
		tx, sign := serialize.Decode36(msg.Payload, msg.LenPayload)
		if sh.VerifyShardTransaction(tx, sign, 4) {
			Shardtxch <- tx
			log.Println("Recieved sharding tx from", msg.Sender)
			send <- msg
		}
	}
}

// read the messages and handle
func handle(p *p2p.Peer, send chan p2p.Msg, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction, errChan chan p2p.Peer) {
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

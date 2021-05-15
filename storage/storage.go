package storage

import (
	dt "LSDI_SN/DataTypes"
	db "LSDI_SN/database"
	"LSDI_SN/p2p"
	"LSDI_SN/serialize"
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
)

// DB badger database
type DB *badger.DB

var (
	f, _    = os.Create("temp/logFile.txt")
	logLock sync.Mutex
)

var orphanedTransactions = make(map[string][]dt.Vertex)
var mux sync.Mutex

// Server ...
type Server struct {
	ForwardingCh chan p2p.Msg
	ServerCh     chan dt.ForwardTx
	DB           *badger.DB
}

//Hash returns the SHA256 hash value
func Hash(b []byte) []byte {
	h := sha256.Sum256(b)
	return []byte(h[:])
}

//AddTransaction checks if transaction if already present in the dag, if not adds to dag and database and returns true else returns false
func (srv *Server) AddTransaction(tx dt.Transaction, signature []byte) int {
	// change this function for the storage node
	var node dt.Vertex
	var duplicationCheck int
	duplicationCheck = 0
	s := serialize.Encode(tx)
	Txid := Hash(s)
	h := serialize.EncodeToHex(Txid[:])
	s = append(s, signature...)
	if !CheckifPresentDb(srv.DB, Txid) { //Duplication check
		node.Tx = tx
		node.Signature = signature
		var tip [32]byte
		if tx.LeftTip == tip && tx.RightTip == tip {
			duplicationCheck = 1
			fmt.Println("Genesis here")
			db.AddToDb(srv.DB, Txid, s)
		} else {
			left := serialize.EncodeToHex(tx.LeftTip[:])
			right := serialize.EncodeToHex(tx.RightTip[:])
			okL := CheckifPresentDb(srv.DB, tx.LeftTip[:])
			okR := CheckifPresentDb(srv.DB, tx.RightTip[:])
			if !okR || !okL {
				if !okL {
					duplicateOrphanTx := false
					log.Println("left orphan transaction")
					log.Println(left)
					mux.Lock()
					for _, orphantx := range orphanedTransactions[left] {
						if bytes.Compare(orphantx.Signature, node.Signature) == 0 {
							duplicateOrphanTx = true
							break
						}
					}
					if !duplicateOrphanTx {
						orphanedTransactions[left] = append(orphanedTransactions[left], node)
					}
					mux.Unlock()
					duplicationCheck = 2
				}
				if !okR {
					duplicateOrphanTx := false
					log.Println("right orphan transaction")
					log.Println(right)
					mux.Lock()
					for _, orphantx := range orphanedTransactions[right] {
						if bytes.Compare(orphantx.Signature, node.Signature) == 0 {
							duplicateOrphanTx = true
							break
						}
					}
					if !duplicateOrphanTx {
						orphanedTransactions[right] = append(orphanedTransactions[right], node)
					}
					mux.Unlock()
					duplicationCheck = 2
				}
			} else {
				duplicationCheck = 1
				db.AddToDb(srv.DB, Txid, s)
			}
		}
	}
	if duplicationCheck == 1 {
		srv.checkorphanedTransactions(h, s)
	}
	return duplicationCheck
}

//GetTransactiondb Wrapper function for GetTransaction in db module
func GetTransactiondb(DB *badger.DB, Txid []byte) (dt.Transaction, []byte) {
	stream := db.GetValue(DB, Txid)
	return serialize.Decode32(stream, uint32(len(stream)))
}

// CheckifPresentDb Wrapper function for CheckKey in db module
func CheckifPresentDb(DB *badger.DB, Txid []byte) bool {
	return db.CheckKey(DB, Txid)
}

//checkorphanedTransactions Checks if any other transaction already arrived has any relation with this transaction, Used in the AddTransaction function
func (srv *Server) checkorphanedTransactions(h string, serializedTx []byte) {
	mux.Lock()
	list, ok := orphanedTransactions[h]
	mux.Unlock()
	if ok {
		for _, node := range list {
			if srv.AddTransaction(node.Tx, node.Signature) == 1 {
				log.Println("resolved Transaction")
			}
		}
	}
	mux.Lock()
	delete(orphanedTransactions, h)
	mux.Unlock()
	return
}

// Run spawns a storage server which spins on a channel (ServerCh) to accept transactions and add them to the DAG.
func (srv *Server) Run() {
	for {
		node := <-srv.ServerCh
		if node.Forward {
			dup := srv.AddTransaction(node.Tx, node.Signature)
			if dup == 1 {
				logLock.Lock()
				f.WriteString(fmt.Sprintf("%s %d %d\n", node.Peer.RemoteAddr().String(), time.Now().Minute(), time.Now().Second()))
				logLock.Unlock()
				var msg p2p.Msg
				msg.ID = 32
				msg.Payload = append(serialize.Encode(node.Tx), node.Signature...)
				msg.LenPayload = uint32(len(msg.Payload))
				srv.ForwardingCh <- msg
			} else if dup == 2 {
				var msg p2p.Msg
				msg.ID = 34
				if !CheckifPresentDb(srv.DB, node.Tx.LeftTip[:]) {
					msg.Payload = node.Tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
				if !CheckifPresentDb(srv.DB, node.Tx.RightTip[:]) {
					msg.Payload = node.Tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
			}
		} else {
			dup := srv.AddTransaction(node.Tx, node.Signature)
			if dup == 2 {
				var msg p2p.Msg
				msg.ID = 34
				if !CheckifPresentDb(srv.DB, node.Tx.LeftTip[:]) {
					msg.Payload = node.Tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
				if !CheckifPresentDb(srv.DB, node.Tx.RightTip[:]) {
					msg.Payload = node.Tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
			}
		}
	}
}

package storage

import (
	dt "Go-DAG-storageNode/DataTypes"
	db "Go-DAG-storageNode/database"
	"Go-DAG-storageNode/serialize"
	"crypto/sha256"
	"log"
	"sync"
)

var orphanedTransactions = make(map[string][]dt.Vertex)
var mux sync.Mutex

//Hash returns the SHA256 hash value
func Hash(b []byte) []byte {
	h := sha256.Sum256(b)
	return []byte(h[:])
}

//AddTransaction checks if transaction if already present in the dag, if not adds to dag and database and returns true else returns false
func AddTransaction(tx dt.Transaction, signature []byte) int {
	// change this function for the storage node
	var node dt.Vertex
	var duplicationCheck int
	duplicationCheck = 0
	s := serialize.Encode(tx)
	Txid := Hash(s)
	h := serialize.EncodeToHex(Txid[:])
	s = append(s, signature...)
	if !CheckifPresentDb(Txid) { //Duplication check
		node.Tx = tx
		node.Signature = signature
		var tip [32]byte
		if tx.LeftTip == tip && tx.RightTip == tip {
			duplicationCheck = 1
			db.AddToDb(Txid, s)
		} else {
			left := serialize.EncodeToHex(tx.LeftTip[:])
			right := serialize.EncodeToHex(tx.RightTip[:])
			okL := CheckifPresentDb(tx.LeftTip[:])
			okR := CheckifPresentDb(tx.RightTip[:])

			if !okL || !okR {
				log.Println("Orphan Transaction")
				if !okL {
					orphanedTransactions[left] = append(orphanedTransactions[left], node)
				}
				if !okR {
					orphanedTransactions[right] = append(orphanedTransactions[right], node)
				}
				duplicationCheck = 2
			} else {
				duplicationCheck = 1
				db.AddToDb(Txid, s)
			}
		}
	}
	if duplicationCheck == 1 {
		checkorphanedTransactions(h, s)
	}
	return duplicationCheck
}

//GetTransactiondb Wrapper function for GetTransaction in db module
func GetTransactiondb(Txid []byte) (dt.Transaction, []byte) {
	stream := db.GetValue(Txid)
	return serialize.Decode32(stream, uint32(len(stream)))
}

// CheckifPresentDb Wrapper function for CheckKey in db module
func CheckifPresentDb(Txid []byte) bool {
	return db.CheckKey(Txid)
}

//GetAllHashes Wrapper function for GetAllKeys in db module
func GetAllHashes() [][]byte {
	return db.GetAllKeys()
}

//checkorphanedTransactions Checks if any other transaction already arrived has any relation with this transaction, Used in the AddTransaction function
func checkorphanedTransactions(h string, serializedTx []byte) {
	mux.Lock()
	list, ok := orphanedTransactions[h]
	mux.Unlock()
	if ok {
		for _, node := range list {
			if AddTransaction(node.Tx, node.Signature) == 1 {
				log.Println("resolved Transaction")
			}
		}
	}
	mux.Lock()
	delete(orphanedTransactions, h)
	mux.Unlock()
	return
}

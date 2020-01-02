package storage

import(
	"fmt"
	dt "Go-DAG-storageNode/datatypes"
	"Go-DAG-storageNode/sync"
	"Go-DAG-storageNode/serialize"
	db "Go-DAG-storageNode/database"
	"net"
	"sync"
)

var OrphanedTransactions = make(map[string] []Vertex)
var Mux sync.Mutex

//Hash returns the SHA256 hash value
func Hash(b []byte) [32]byte {
	h := sha256.Sum256(b)
	return h
}



func AddTransaction(tx dt.Transaction, signature []byte, serializedTx []byte, Peers *dt.Peers) bool {

	// change this function for the storage Node

	Txid := Hash(serialize.SerializeData(tx))
	// h := serialize.EncodeToHex(Txid[:])

	if tx.Timestamp == 0 { 						//To add the genesis tx to the db
		db.AddToDb(Txid,serializedTx)
		return true
	}
	var Vertex dt.Vertex
	var duplicationCheck bool
	duplicationCheck = false
	
	if !db.checkifPresentDb(Txid){  //Duplication check
		Vertex.Tx = tx
		Vertex.Signature = signature
		left := serialize.EncodeToHex(tx.LeftTip[:])
		right := serialize.EncodeToHex(tx.RightTip[:]) 
		okL := db.checkifPresentDb(left)
		okR := db.checkifPresentDb(right)
		if !okL || !okR {
			if !okL {
				// OrphanedTransactions[left] = append(OrphanedTransactions[left],Vertex)
				//fmt.Println("Orphaned Transactions")
				sync.QueryOneTransactions(Peers,&okL)
			}
			if !okR {
				sync.QueryOneTransactions(Peers,&okR)
				//fmt.Println("Orphaned Transactions")
				// OrphanedTransactions[right] = append(OrphanedTransactions[right],Vertex)
			}
			if okL && okR {
				return false
			}
		} else {
			// l := getTx(left)
			// r := getTx(right)
			// serializedTx := serialize.EncodeToHex(serialize.SerializeData(tx))
			// Txid := tx.Txid
			db.AddToDb(Txid,serializedTx)
			// if left == right {
			// 	l.Neighbours = append(l.Neighbours,h)
			// 	dag.Graph[serialize.EncodeToHex(tx.LeftTip[:])] = l
			// } else {
			// 	l.Neighbours = append(l.Neighbours,h)
			// 	dag.Graph[serialize.EncodeToHex(tx.LeftTip[:])] = l
			// 	r.Neighbours = append(r.Neighbours,h)
			// 	dag.Graph[serialize.EncodeToHex(tx.RightTip[:])] = r
			// }
			duplicationCheck = true
			// //fmt.Println("Added Transaction ",h)
		}
	}
	// if duplicationCheck {
	// 	checkOrphanedTransactions(h, serializedTx)
	// }
	return duplicationCheck
}

//GetTransactiondb Wrapper function for GetTransaction in db module
func GetTransactiondb(Txid []byte) dt.Transaction,[]byte {
	stream := db.GetValue(Txid)
	return serialize.DeserializeTransaction(stream)
}

//checkifPresentDb Wrapper function for CheckKey in db module
func checkifPresentDb(Txid []byte) bool {
	return db.CheckKey(Txid)
}

//GetAllHashes Wrapper function for GetAllKeys in db module
func GetAllHashes() []string{
	return db.GetAllKeys()
}
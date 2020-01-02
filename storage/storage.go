package storage

import(
	// "fmt"
	dt "Go-DAG-storageNode/DataTypes"
	// "Go-DAG-storageNode/sync"
	"Go-DAG-storageNode/serialize"
	db "Go-DAG-storageNode/database"
	// "net"
	"time"
	"math/rand"
	"crypto/sha256"
	"sync"
)

var OrphanedTransactions = make(map[string] [] dt.Vertex)
var Mux sync.Mutex

//Hash returns the SHA256 hash value
func Hash(b []byte) []byte {
	h := sha256.Sum256(b)
	return []byte(h[:])
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
	
	if !db.CheckKey(Txid){  //Duplication check
		Vertex.Tx = tx
		Vertex.Signature = signature
		left := tx.LeftTip[:]
		right := tx.RightTip[:]
		okL := db.CheckKey(left)
		okR := db.CheckKey(right)
		if !okL || !okR {
			if !okL {
				// OrphanedTransactions[left] = append(OrphanedTransactions[left],Vertex)
				//fmt.Println("Orphaned Transactions")
				QueryOneTransactions(Peers,Txid,&okL)
			}
			if !okR {
				QueryOneTransactions(Peers,Txid,&okR)
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
func GetTransactiondb(Txid []byte) (dt.Transaction,[]byte) {
	stream := db.GetValue(Txid)
	return serialize.DeserializeTransaction(stream)
}

//checkifPresentDb Wrapper function for CheckKey in db module
func checkifPresentDb(Txid []byte) bool {
	return db.CheckKey(Txid)
}

//GetAllHashes Wrapper function for GetAllKeys in db module
func GetAllHashes() [][]byte{
	return db.GetAllKeys()
}

func QueryOneTransactions(Peers *dt.Peers, hash []byte, valid *bool) {
	var magicNumber uint32
	magicNumber = 3
	bytes := serialize.EncodeToBytes(magicNumber)
	// b,_ := hex.DecodeString(hash)
	b := append(bytes,hash...)
	rand.Seed(time.Now().UnixNano())
    p := rand.Perm(len(Peers.Fds))
    var keys []string
    for k := range Peers.Fds {
    	keys = append(keys, k)
	}
    *valid = false
    for _, r := range p {
    	Peers.Mux.Lock()
    	Peers.Fds[keys[r]].Write(b)
    	bufCheck := make([]byte,1)
		_,err := Peers.Fds[keys[r]].Read(bufCheck)
		if serialize.EncodeToHex(bufCheck) == "1" {
	    	buf := make([]byte,1024)
			length,_ := Peers.Fds[keys[r]].Read(buf)
			Peers.Mux.Unlock()
			tx,sign := serialize.DeserializeTransaction(buf[:length])
			AddTransaction(tx,sign,buf[:length],Peers)
			*valid = true
			break
		}
		if err != nil {
			panic(err)
		}
		Peers.Mux.Unlock()
    }
	//Verify signature b4 adding to db - NOT DOING
	
}
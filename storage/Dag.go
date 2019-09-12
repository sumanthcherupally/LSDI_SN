package storage

import(
	dt "Go-DAG-storageNode/DataTypes"
	"fmt"
	"Go-DAG-storageNode/serialize"
	"Go-DAG-storageNode/Crypto"
	"sync"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var OrphanedTransactions = make(map[string]dt.Node)
var Mux sync.Mutex 


func AddTransaction(dag *dt.DAG,tx dt.Transaction, signature []byte) bool {

	// change this function for the storage node
	var node dt.Node
	var duplicationCheck bool
	duplicationCheck = false
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])

	dag.Mux.Lock()
	if !checkifPresentDb(h){  //Duplication check
		node.Tx = tx
		node.Signature = signature
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:]) 
		ok_l := checkifPresentDb(left)
		ok_r := checkifPresentDb(right)
		if !ok_l || !ok_r {
			if !ok_l {
				OrphanedTransactions[left] = node
				fmt.Println("Orphaned Transactions")	
			}
			if !ok_r {
				fmt.Println("Orphaned Transactions")
				OrphanedTransactions[right] = node
			}
		} else {
			// l := getTx(left)
			// r := getTx(right)
			serializedTx := serialize.EncodeToBytes(tx)
			tips := left+","+right
			Txid := tx.Txid
			AddToDb(serializedTx,Txid,h,tips,signature)
			// if left == right {
			// 	l.Neighbours = append(l.Neighbours,h)
			// 	dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			// } else {
			// 	l.Neighbours = append(l.Neighbours,h)
			// 	dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			// 	r.Neighbours = append(r.Neighbours,h)
			// 	dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])] = r
			// }
			duplicationCheck = true
			// //fmt.Println("Added Transaction ",h)
		}
	}
	dag.Mux.Unlock()
	if duplicationCheck {
		checkOrphanedTransactions(h,dag)
	}
	return duplicationCheck
}

func AddToDb(serializedTx []byte, Txid [16]byte, h string,tips string,signature []byte) {
	db, err := sql.Open("mysql",
		"root:sumanth@tcp(127.0.0.1:3306)/dag")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO storage(Hash_tx,Txid,Transaction,Tips,Signature) VALUES(?,?,?,?,?)")
	res, err := stmt.Exec(h,Txid,serializedTx,tips,signature)
	if err != nil {
	log.Fatal(err)
	}
	fmt.Println(res)
}


func checkifPresentDb(h string) bool{
	db, err := sql.Open("mysql","root:sumanth@tcp(127.0.0.1:3306)/dag")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	stmt, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM storage WHERE Hash_tx = ?)")
	res, err := stmt.Query(h)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()
	var present int
	err1 := res.Scan(&present)
	if err1 != nil {
		log.Fatal(err1)
	}
	if(present==1){
		return true
	} else{
		return false
	}
}

func checkOrphanedTransactions(h string,dag *dt.DAG) {
	Mux.Lock()
	node,ok := OrphanedTransactions[h]
	Mux.Unlock()
	if ok {
		if AddTransaction(dag,node.Tx,node.Signature) {
			//fmt.Println("resolved Transaction")
		}
	}
	Mux.Lock()
	delete(OrphanedTransactions,h)
	Mux.Unlock()
	return 
}

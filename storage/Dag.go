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


var OrphanedTransactions = make(map[string]dt.Vertex)
var Mux sync.Mutex 


func AddTransaction(tx dt.Transaction, signature []byte) bool {

	// change this function for the storage Node
	if tx.Timestamp == 0 {
		serializedTx := Crypto.EncodeToHex(serialize.SerializeData(tx))
		Txid := [16]byte(tx.Txid)
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:])
		tips := left+","+right
		hash := Crypto.Hash(serialize.SerializeData(tx))
		h := Crypto.EncodeToHex(hash[:])
		AddToDb(serializedTx,Txid,h,tips,signature)
		return true
	}
	var Vertex dt.Vertex
	var duplicationCheck bool
	duplicationCheck = false
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])

	if !checkifPresentDb(h){  //Duplication check
		Vertex.Tx = tx
		Vertex.Signature = signature
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:]) 
		ok_l := checkifPresentDb(left)
		ok_r := checkifPresentDb(right)
		if !ok_l || !ok_r {
			if !ok_l {
				OrphanedTransactions[left] = Vertex
				fmt.Println("Orphaned Transactions")	
			}
			if !ok_r {
				fmt.Println("Orphaned Transactions")
				OrphanedTransactions[right] = Vertex
			}
		} else {
			// l := getTx(left)
			// r := getTx(right)
			serializedTx := Crypto.EncodeToHex(serialize.SerializeData(tx))
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
	if duplicationCheck {
		checkOrphanedTransactions(h)
	}
	return duplicationCheck
}

func AddToDb(serializedTx string, Txid [16]byte, h string,tips string,signature []byte) {
	db, err := sql.Open("mysql","root:sumanth@tcp(127.0.0.1:3306)/dag")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO storage(Hash_tx,Txid,Transaction,Tips,Signature) VALUES(?,?,?,?,?)")
	_, err = stmt.Exec(h,Crypto.EncodeToHex(Txid[:]),serializedTx,tips,signature)
	if err != nil {
	log.Fatal(err)
	}
}

func GetAllHashes() []string {
	db, err := sql.Open("mysql","root:sumanth@tcp(127.0.0.1:3306)/dag")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	stmt, err := db.Prepare("SELECT Hash_tx FROM storage")
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	Hashes := make([]string,0)
	for rows.Next() {
		var u string
		err := rows.Scan(&u) // check err
		if err != nil {
			log.Fatal(err)
		}
		Hashes = append(Hashes, u)
	}
	return Hashes
}

func GetTransaction(hash string) dt.Transaction {
	db, err := sql.Open("mysql","root:sumanth@tcp(127.0.0.1:3306)/dag")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var Resp string
	queryStr := `SELECT Transaction FROM storage WHERE Hash_tx = ?` // check err
	err1 := db.QueryRow(queryStr, hash).Scan(&Resp)
	if err1 != nil {
		log.Fatal(err1)
	}
	return serialize.Deserializedata(Crypto.DecodeToBytes(Resp))
}

func GetSignature(hash string) []byte {
	db, err := sql.Open("mysql","root:sumanth@tcp(127.0.0.1:3306)/dag")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	Resp := make([]byte,0)
	queryStr := `SELECT Signature FROM storage WHERE Hash_tx = ?` // check err
	err1 := db.QueryRow(queryStr, hash).Scan(&Resp)
	if err1 != nil {
		log.Fatal(err1)
	}
	return Resp
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
	res.Next()
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

func checkOrphanedTransactions(h string) {
	Mux.Lock()
	Vertex,ok := OrphanedTransactions[h]
	Mux.Unlock()
	if ok {
		if AddTransaction(Vertex.Tx,Vertex.Signature) {
			fmt.Println("resolved Transaction")
		}
		Mux.Lock()
		delete(OrphanedTransactions,h)
		Mux.Unlock()
	}
	return 
}

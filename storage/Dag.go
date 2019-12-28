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
	badger "github.com/dgraph-io/badger"
)


var OrphanedTransactions = make(map[string] []dt.Vertex)
var Mux sync.Mutex 


func AddTransaction(tx dt.Transaction, signature []byte, serializedTx []byte) bool {

	// change this function for the storage Node

	Txid := Crypto.Hash(serialize.SerializeData(tx))
	h := Crypto.EncodeToHex(Txid[:])

	if tx.Timestamp == 0 { 						//To add the genesis tx to the db
		// serializedTx := Crypto.EncodeToHex(serialize.SerializeData(tx))
		// Txid := [16]byte(tx.Txid)
		// left := Crypto.EncodeToHex(tx.LeftTip[:])
		// right := Crypto.EncodeToHex(tx.RightTip[:])	
		// AddToDb(Txid,serializedTx,h,left,right,signature)
		AddToDb(Txid,serializedTx)
		return true
	}
	var Vertex dt.Vertex
	var duplicationCheck bool
	duplicationCheck = false
	
	if !checkifPresentDb(Txid){  //Duplication check
		Vertex.Tx = tx
		Vertex.Signature = signature
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:]) 
		okL := checkifPresentDb(left)
		okR := checkifPresentDb(right)
		if !okL || !okR {
			if !okL {
				OrphanedTransactions[left] = append(OrphanedTransactions[left],Vertex)
				//fmt.Println("Orphaned Transactions")	
			}
			if !okR {
				//fmt.Println("Orphaned Transactions")
				OrphanedTransactions[right] = append(OrphanedTransactions[right],Vertex)
			}
		} else {
			// l := getTx(left)
			// r := getTx(right)
			// serializedTx := Crypto.EncodeToHex(serialize.SerializeData(tx))
			// Txid := tx.Txid
			AddToDb(Txid,serializedTx)
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
		checkOrphanedTransactions(h, serializedTx)
	}
	return duplicationCheck
}

func AddToDb(Txid []byte, serializedTx []byte]) {	
	opts := badger.DefaultOptions
	opts.Dir = ""
	opts.ValueDir = ""
	kv, err := badger.NewKV(&opts)
	if err != nil {
		panic(err)
	}
	defer kv.Close()
	err = kv.Set(Txid,serializedTx)
	if err != nil {
		panic(err)
	}

	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO storage(Hash_tx,Txid,Transaction,Left_tip,Right_tip,Signature) VALUES(?,?,?,?,?,?)")
	// _, err = stmt.Exec(h,Crypto.EncodeToHex(Txid[:]),serializedTx,left,right,signature)
	// if err != nil {
	// log.Fatal(err)
	// }
}


func GetAllHashes() []string {
	Hashes := make([]string,0)
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Dir = ""
		// opts.ValueDir = ""
		opts.PrefetchSize = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			Hashes = append(Hashes, it.Item().Key())
		}
		return nil
	})
	// fmt.Println(len(Hashes))
	return Hashes

	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()
	// stmt, err := db.Prepare("SELECT Hash_tx FROM storage")
	// rows, err := stmt.Query()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()
	// Hashes := make([]string,0)
	// for rows.Next() {
	// 	var u string
	// 	err := rows.Scan(&u) // check err
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	Hashes = append(Hashes, u)
	// }
	// fmt.Println(len(Hashes))
	// return Hashes
}

// GetTransaction returns transaction based on hash value.
func GetTransaction(Txid []byte]) dt.Transaction {
	
	opts := badger.DefaultOptions
	opts.Dir = ""
	opts.ValueDir = ""
	kv, err := badger.NewKV(&opts)
	if err != nil {
		panic(err)
	}
	defer kv.Close()
	var item badger.KVItem
	err = kv.Get(Txid,&item)
	if err == ErrKeyNotFound {
		return false
	}
	return serialize.Deserializedata(item.Value())


	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()
	// var Resp string
	// queryStr := "SELECT Transaction FROM storage WHERE Hash_tx = ?" // check err
	// err1 := db.QueryRow(queryStr, hash).Scan(&Resp)
	// if err1 != nil {
	// 	log.Fatal(err1)
	// }
	// return serialize.Deserializedata(Crypto.DecodeToBytes(Resp))
}

// GetSignature returns signature of tranasction based on hash value.
// func GetSignature(hash string) []byte {
// 	opts := badger.DefaultOptions
// 	opts.Dir = ""
// 	opts.ValueDir = ""
// 	kv, err := badger.NewKV(&opts)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer kv.Close()
// 	var item badger.KVItem
// 	err = kv.Get([]byte(h),&item)
// 	if err == ErrKeyNotFound {
// 		return false
// 	}


// 	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// defer db.Close()
// 	// Resp := make([]byte,0)
// 	// queryStr := `SELECT Signature FROM storage WHERE Hash_tx = ?` // check err
// 	// err1 := db.QueryRow(queryStr, hash).Scan(&Resp)
// 	// if err1 != nil {
// 	// 	log.Fatal(err1)
// 	// }
// 	// return Resp
// }

func checkifPresentDb(Txid []byte]) bool{

	opts := badger.DefaultOptions
	opts.Dir = ""
	opts.ValueDir = ""
	kv, err := badger.NewKV(&opts)
	if err != nil {
		panic(err)
	}
	defer kv.Close()
	var item badger.KVItem
	err = kv.Get(Txid,&item)
	if err == ErrKeyNotFound {
		return false
	} else if err != nil {
		panic(err)
	} else {
		return true
	}

	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()
	// stmt, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM storage WHERE Hash_tx = ?)")
	// res, err := stmt.Query(h)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer res.Close()
	// var present int
	// res.Next()
	// err1 := res.Scan(&present)
	// if err1 != nil {
	// 	log.Fatal(err1)
	// }
	// if(present==1){
	// 	return true
	// } else{
	// 	return false
	// }
}

func checkOrphanedTransactions(h string, serializedTx []byte) {
	Mux.Lock()
	vertices,ok := OrphanedTransactions[h]
	Mux.Unlock()
	if ok {
		for _,vertex := range vertices {
			if AddTransaction(vertex.Tx,vertex.Signature,serializedTx) {
				//fmt.Println("resolved transaction")
			}
		}
		Mux.Lock()
		delete(OrphanedTransactions,h)
		Mux.Unlock()
	}
	return 
}

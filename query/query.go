package query

import (
    "fmt"
    "log"
	"net/http"
	"encoding/json"
	"database/sql"
	"Go-DAG-storageNode/serialize"
	_ "github.com/go-sql-driver/mysql"
	"Go-DAG-storageNode/Crypto"
)

type verifyQuery struct{
	Txid string
}

type verifyQueryResp struct{
	tx string
}

type send struct{
	Hash [32]byte
}

// type dbResponse struct {
// 	Hash_tx string
// 	Txid [16]byte
// 	Transaction []byte
// 	Tips string
// 	Signature []byte
// }

func HandleQuery(w http.ResponseWriter, r *http.Request) {
	query := verifyQuery{}
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil{
		panic(err)
	}

	db, err := sql.Open("mysql","root:sumanth@tcp(127.0.0.1:3306)/dag")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	Resp := verifyQueryResp{}
	queryStr := `SELECT Transaction FROM storage WHERE Txid = ?` // check err
	err1 := db.QueryRow(queryStr, query.Txid).Scan(&Resp.tx)
	if err1 != nil {
		log.Fatal(err1)
	}
	RespToSend := serialize.Deserializedata(Crypto.DecodeToBytes(Resp.tx))
	qq := send{}
	qq.Hash = RespToSend.Hash
	ww ,err := json.Marshal(qq)
	if err != nil{
		panic(err)
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ww)
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/query",HandleQuery)
 
    fmt.Printf("Started server for querying HTTP POST...\n")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal(err)
    }
}
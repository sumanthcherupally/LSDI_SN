package query

import (
	"DAG-SN/Crypto"
	"DAG-SN/storage"
	"encoding/json"
	"log"
	"net/http"
)

type verifyQuery struct {
	Txid string
}

type verifyQueryResp struct {
	tx string
}

type send struct {
	Hash string
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	query := verifyQuery{}
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		panic(err)
	}

	Resp := verifyQuery{}
	tx, _ := storage.GetTransactiondb([]byte(Resp.Txid))
	qq := send{}
	qq.Hash = Crypto.EncodeToHex(tx.Hash[:])
	ww, err := json.Marshal(qq)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ww)
}

// StartServer serves the http requests for transactions
func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/query", handleQuery)
	//Log.Printf("Started server for querying HTTP POST...\n")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

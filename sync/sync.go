package sync

import (
	"net"
	"Go-DAG-storageNode/serialize"
	"Go-DAG-storageNode/storage"
	"encoding/json"
	"encoding/hex"
)

// RequestHashes requests all the hashes of Transactions in the DAG
// These hashes can be used to query the transactions  
func RequestHashes(conn net.Conn) []string{
	var magicNumber uint32
	magicNumber = 2
	bytes := serialize.EncodeToBytes(magicNumber)
	conn.Write(bytes)
	var serialMsg []byte
	for {
		buf := make([]byte,1024)
		length,_ := conn.Read(buf)
		serialMsg = append(serialMsg,buf...)
		if length < 1024 {
			break
		}
	}
	var hashes []string
	json.Unmarshal(serialMsg,&hashes)
	return hashes
}

// QueryTransactions gets the Transactions based on hashes
func QueryTransactions(conn net.Conn, hashes []string) {
	var magicNumber uint32
	magicNumber = 3
	bytes := serialize.EncodeToBytes(magicNumber)
	for _,v := range hashes {
		b,_ := hex.DecodeString(v)
		b = append(bytes,b...)
		conn.Write(b)
		buf := make([]byte,1024)
		length,_ := conn.Read(buf)
		tx,sign = serialize.DeserializeTransaction(buf[:l])
		storage.AddTransaction(tx,sign)
	}
}
package sync

import (
	"net"
	"Go-DAG-storageNode/serialize"
	"Go-DAG-storageNode/storage"
	"encoding/json"
	"encoding/hex"
	"fmt"
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
		if length < 1024 {
			serialMsg = append(serialMsg,buf[:length]...)
			break
		} else{
			serialMsg = append(serialMsg,buf...)
		}
	}
	var hashes []string
	err := json.Unmarshal(serialMsg,&hashes)
	if err!=nil{
		fmt.Println(err)
	}
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
		tx,sign := serialize.DeserializeTransaction(buf[:length])
		//Verify signature b4 adding to db - NOT DOING
		storage.AddTransaction(tx,sign,buf[:length])
	}

	// missingTxs := requestMissingTransacions()
	// fmt.Println(len(missingTxs))

	// for _,v := range missingTxs {
	// 	b,_ := hex.DecodeString(v)
	// 	b = append(bytes,b...)
	// 	conn.Write(b)
	// 	buf := make([]byte,1024)
	// 	length,_ := conn.Read(buf)
	// 	tx,sign := serialize.DeserializeTransaction(buf[:length])
	// 	storage.AddTransaction(tx,sign)
	// }
}

func QueryOneTransactions(Peers *dt.Peers, hash string, valid &bool) bool {
	var magicNumber uint32
	magicNumber = 3
	bytes := serialize.EncodeToBytes(magicNumber)
	b,_ := hex.DecodeString(hash)
	b = append(bytes,b...)
	rand.Seed(time.Now().UnixNano())
    p := rand.Perm(len(Peers.Fds))
    valid = false
    for _, r := range p {
    	Peers.Mux.Lock()
    	Peers.Fds[r].Write(b)
    	bufCheck := make([]byte,1)
		l,_ := Peers.Fds[r].Read(bufCheck)
		if Crypto.EncodeToHex(bufCheck) == "1" {
	    	buf := make([]byte,1024)
			length,_ := Peers.Fds[r].Read(buf)
			Peers.Mux.Unlock()
			tx,sign := serialize.DeserializeTransaction(buf[:length])
			storage.AddTransaction(tx,sign,buf[:length])
			valid = true
			break
		}
		Peers.Mux.Unlock()
    }
	//Verify signature b4 adding to db - NOT DOING
	
}
/*
func requestMissingTransacions() ([]string){

	var missingTransactions []string
	
	for _,node := range storage.OrphanedTransactions {
		tx := node.Tx
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:])

		_,okL := storage.OrphanedTransactions[left]
		_,okR := storage.OrphanedTransactions[right]

		if !okL {
			missingTransactions = append(missingTransactions,left)
		}
		if !okR {
			missingTransactions = append(missingTransactions,right)
		}
	}

	return missingTransactions

}
*/
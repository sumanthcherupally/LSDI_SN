package sharding

import (
	"LSDI_SN/Crypto"
	dt "LSDI_SN/DataTypes"
	pow "LSDI_SN/Pow"
	"LSDI_SN/serialize"
	"log"
	"time"
)

//VerifyDiscovery verifies shard signal from discovery
func VerifyDiscovery(msg dt.ShardSignal, signature []byte) bool {
	//Deserialise the message
	//Verify signature using PuK
	//Return valid or not
	return true
}

//VerifyShardTransaction verifies shard transaction
func VerifyShardTransaction(tx dt.ShardTransaction, signature []byte, difficulty int) bool {
	//Verify Puk
	//Verify signature
	//Verify Pow
	//Verify Shard number
	s := serialize.Encode(tx)
	SerialKey := tx.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	h := Crypto.Hash(s)
	sigVerify := Crypto.Verify(signature, PublicKey, h[:])
	if sigVerify == false {
		log.Println("INVALID SIGNATURE")
	}
	return sigVerify && pow.VerifyPoW(tx, difficulty) && tx.ShardNo == tx.Nonce%2
}

//MakeShardingtx Call on recieving sharding signal from discovery after forwarding it
func MakeShardingtx(Puk []byte, Signal dt.ShardSignal) (dt.ShardTransaction, error) {
	difficulty := 4
	var tx dt.ShardTransaction
	//Create transaction
	copy(tx.From[:], Puk)
	tx.Timestamp = time.Now().UnixNano()
	tx.Nonce = 0
	tx.ShardNo = 0
	tx.Identifier = Signal.Identifier
	//Do PoW
	pow.PoW(&tx, difficulty)
	//Wait for recieving messages
	//	//Verify each recvd pow and add to buffer
	//	//Wait till threshold or timeout
	//Update peers
	return tx, nil
}

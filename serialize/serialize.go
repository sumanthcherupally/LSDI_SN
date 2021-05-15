package serialize

import (
	dt "LSDI_SN/DataTypes"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"reflect"
)

//EncodeToHex converts byte slice to the string
func EncodeToHex(data []byte) string {
	return hex.EncodeToString(data)
}

//DecodeToBytes converts string to byte slice
func DecodeToBytes(data string) []byte {
	b, err := hex.DecodeString(data)
	if err != nil {
		log.Print(err)
	}
	return b
}

//Encode serializes transaction to byte slice
func Encode(t interface{}) []byte {
	// iterating over a struct is painful in golang
	var b []byte
	switch tx := t.(type) {
	case dt.Transaction:
		b = append(b, EncodeToBytes(tx.Timestamp)...)
		b = append(b, EncodeToBytes(tx.Hash)...)
		b = append(b, EncodeToBytes(tx.From)...)
		b = append(b, EncodeToBytes(tx.LeftTip)...)
		b = append(b, EncodeToBytes(tx.RightTip)...)
		b = append(b, EncodeToBytes(tx.Nonce)...)
	case dt.ShardTransaction:
		v := reflect.ValueOf(&tx).Elem()
		for i := 0; i < v.NumField(); i++ {
			value := v.Field(i)
			b = append(b, EncodeToBytes(value.Interface())...)
		}
	case dt.ShardSignal:
		v := reflect.ValueOf(&tx).Elem()
		for i := 0; i < v.NumField(); i++ {
			value := v.Field(i)
			b = append(b, EncodeToBytes(value.Interface())...)
		}
	}
	return b
}

//EncodeToBytes Converts any type to byte slice, supports strings,integers etc.
func EncodeToBytes(x interface{}) []byte {
	// encode based on type as binary package doesn't support strings
	switch x.(type) {
	case string:
		str := x.(string)
		return []byte(str)
	default:
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, x)
		return buf.Bytes()
	}
}

//Decode32 Converts back byte slice to transaction
func Decode32(payload []byte, lenPayload uint32) (dt.Transaction, []byte) {
	// signature := payload[lenPayload-72:]
	signature := make([]byte, 72)
	copy(signature, payload[lenPayload-72:])
	r := bytes.NewReader(payload[:lenPayload-72])
	var tx dt.Transaction
	err := binary.Read(r, binary.LittleEndian, &tx)
	if err != nil {
		log.Println(err)
	}
	return tx, signature
}

//Decode35 Converts back byte slice to transaction
func Decode35(payload []byte, lenPayload uint32) (dt.ShardSignal, []byte) {
	signature := payload[lenPayload-72:]
	r := bytes.NewReader(payload[:lenPayload-72])
	var tx dt.ShardSignal
	err := binary.Read(r, binary.LittleEndian, &tx)
	if err != nil {
		log.Println(err)
	}
	return tx, signature
}

//Decode36 Converts back byte slice to transaction
func Decode36(payload []byte, lenPayload uint32) (dt.ShardTransaction, []byte) {
	signature := payload[lenPayload-72:]
	r := bytes.NewReader(payload[:lenPayload-72])
	var tx dt.ShardTransaction
	err := binary.Read(r, binary.LittleEndian, &tx)
	if err != nil {
		log.Println(err, "error due to serialization")
	}
	return tx, signature
}

package serialize

import (
	"reflect"
	"bytes"
	"math/big"
	"encoding/asn1"
	"encoding/binary"
	//"encoding/json"
	"GO-DAG/storage"
	"fmt"
)

//EncodeToHex converts byte slice to the string
func EncodeToHex(data []byte) string {
	return hex.EncodeToString(data)
}

//DecodeToBytes converts string to byte slice
func DecodeToBytes(data string) []byte {
	b,_ := hex.DecodeString(data)
	return b
}

//SerializeData serializes transaction to byte slice
func SerializeData(t storage.Transaction) []byte {
	// iterating over a struct is painful in golang
	var b []byte
	v := reflect.ValueOf(&t).Elem()
	for i := 0; i < v.NumField() ;i++ {
		value := v.Field(i)
		b = append(b,EncodeToBytes(value.Interface())...)
	}
	return b
}

//EncodeToBytes Converts any type to byte slice, supports strings,integers etc.
func EncodeToBytes(x interface{}) []byte {
	// encode based on type as binary package doesn't support strings
	switch x.(type) {
	case string :
		str := x.(string)
		return []byte(str)
	default :
		buf := new(bytes.Buffer)
		binary.Write(buf,binary.LittleEndian, x)
		return buf.Bytes()
	}
}

//DeserializeTransaction Converts back byte slice to transaction
func DeserializeTransaction(b []byte) (storage.Transaction,[]byte) {
	// only a temporary method will change to include signature and other checks
	l := binary.LittleEndian.Uint32(b[:4])
	payload := b[4:l+4]
	signature := b[l+4:]
	r := bytes.NewReader(payload)
	var tx storage.Transaction
	err := binary.Read(r,binary.LittleEndian,&tx)
	if err != nil { 
		fmt.Println(err)
	}
	return tx,signature
}

//canonicalizeInt Converts bigint to byte slice
func canonicalizeInt(val *big.Int) []byte {
	b := val.Bytes()
	if len(b) == 0 {
		b = []byte{0x00}
	}
	if b[0]&0x80 != 0 {
		paddedBytes := make([]byte, len(b)+1)
		copy(paddedBytes[1:], b)
		b = paddedBytes
	}
	return b
}

package server

import(
	"net"
	"time"
	"encoding/binary"
	dt "Go-DAG-storageNode/DataTypes"
	"Go-DAG-storageNode/Crypto"
	"Go-DAG-storageNode/serialize"
	"Go-DAG-storageNode/storage"
	"encoding/json"
	"strings"
	"sync"
	log "Go-DAG-storageNode/logdump"
)

type Server struct {
	Peers *dt.Peers
}

func (srv *Server) HandleConnection(connection net.Conn,dbLock *sync.Mutex) {
	// each connection is handled in a seperate go routine
	addr := connection.RemoteAddr().String()
	ip := addr[:strings.IndexByte(addr,':')]
	srv.Peers.Mux.Lock()
	if _,ok := srv.Peers.Fds[ip] ; !ok {
		c,e := net.Dial("tcp",ip+":9000")
		if e != nil {
			log.Println("CONNECTION UNSUCCESSFUL")
		} else {
			srv.Peers.Fds[ip] = c
			log.Println("CONNECTION WITH PEER SUCCESSFUL")
		}
	}
	srv.Peers.Mux.Unlock()
	for {
		var buf []byte
		buf1 := make([]byte,8) // reading the header 
		headerLen,err := connection.Read(buf1)
		magic_number := binary.LittleEndian.Uint32(buf1[:4]) 
		// specifies the type of the message
		if magic_number == 1 { 
			if headerLen < 8 {
				log.Println("message broken")
			} else {
				length := binary.LittleEndian.Uint32(buf1[4:8])
				buf2 := make([]byte,length+72)
				l,_ := connection.Read(buf2)
				buf = append(buf1,buf2[:l]...)
			}
		} else if magic_number == 2 {
			buf = buf1
		} else if magic_number == 3 {
			buf2 := make([]byte,36)
			l,_ := connection.Read(buf2)
			buf = append(buf1,buf2[:l]...)
		}
		if err != nil {
			// Remove from the list of the peer
			srv.Peers.Mux.Lock()
			delete(srv.Peers.Fds,ip)
			srv.Peers.Mux.Unlock()
			//log.Println(err)
			break
		}
		if len(buf) > 0 {
			srv.HandleRequests(connection,buf,ip,dbLock)
		}
	}
	defer connection.Close()
}


func (srv *Server)HandleRequests (connection net.Conn,data []byte, IP string, dbLock *sync.Mutex) {
	magic_number := binary.LittleEndian.Uint32(data[:4])
	if magic_number == 1 {
		tx,sign := serialize.DeserializeTransaction(data[4:])
		if ValidTransaction(tx,sign) { 
			// maybe wasting verifying duplicate transactions, 
			// instead verify signatures and PoW while tip selection
			dbLock.Lock()
			added := storage.AddTransaction(tx,sign,data[4:])
			dbLock.Unlock()
			if added {
				// log.Println("RECIEVED TRANSACTION")
				// time.Sleep(1*time.Second)
				// log.Println("VERIFIED TRANSACTION PoW AND SIGNATURE")
				// time.Sleep(1*time.Second)
				// log.Println("ADDED TO DATABASE")
				srv.ForwardTransaction(data,IP)
				// log.Println("FORWARDED TO OTHER PEERS")
			}
			
		}
	} else if magic_number == 2 {
		// request to give the hashes of tips 
		ser,_ := json.Marshal(storage.GetAllHashes())
		connection.Write(ser)
	} else if magic_number == 3 {
		// request to give transactions based on tips
		hash := data[4:36]
		str := Crypto.EncodeToHex(hash)
		tx,sign := storage.GetTransaction([]byte(str))
		// sign := storage.GetSignature(str)
		reply := serialize.SerializeData(tx)
		var l uint32
		l = uint32(len(reply))
		reply = append(reply,sign...)
		reply = append(serialize.EncodeToBytes(l),reply...)
		connection.Write(reply)
	} else {
		log.Println("FAILED REQUEST")
	}
}


func ValidTransaction(t dt.Transaction, signature []byte) bool {
	// check the signature
	s := serialize.SerializeData(t)
	SerialKey := t.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	h := Crypto.Hash(s)
	sigVerify := Crypto.Verify(signature,PublicKey,h[:])
	if sigVerify == false {
		log.Println("INVALID SIGNATURE")
	}
	return sigVerify && Crypto.VerifyPoW(t,2)
	//return Crypto.VerifyPoW(t,2)
}


func (srv *Server)ForwardTransaction(t []byte, IP string) {
	// sending the transaction to the peers excluding the one it came from
	//log.Println("Relayed to other peers")	
	srv.Peers.Mux.Lock()
	for k,conn := range srv.Peers.Fds {
		if k != IP {
			conn.Write(t)
		} 
	}
	srv.Peers.Mux.Unlock()
}


func (srv *Server)StartServer() {
	listener, _ := net.Listen("tcp",":9000")
	var dbLock sync.Mutex
	for {
		conn, _ := listener.Accept()
 		go srv.HandleConnection(conn,&dbLock) 
 		// go routine executes concurrently
	}
	defer listener.Close()
}
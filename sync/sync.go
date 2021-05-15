package sync

import (
	"LSDI_SN/p2p"
	"encoding/json"
	"log"
)

// Sync copies the DAG from other nodes
func Sync(p p2p.Peer) {
	var msg p2p.Msg
	msg.ID = 33
	p.Send(msg)
	replyMsg, _ := p.GetMsg()
	var hashes [][]byte
	json.Unmarshal(replyMsg.Payload, &hashes)
	log.Println(len(hashes))
	for _, hash := range hashes {
		msg.ID = 34
		msg.Payload = hash
		msg.LenPayload = uint32(len(msg.Payload))
		p.Send(msg)
	}
}

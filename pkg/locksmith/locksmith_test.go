package locksmith

import (
	"testing"
	"time"

	"github.com/xmliszt/e-safe/pkg/rpc"
	"gopkg.in/go-playground/validator.v9"
)

// Expected a locksmith server to be created with all required fields
func TestInitializeLocksmith(t *testing.T) {
	validate := validator.New()
	locksmith := InitializeLocksmith()
	err := validate.Struct(locksmith)
	if err != nil {
		t.Error(err)
	}
}

// Expected n nodes to be initialized with all required fields
func TestInitializeNodes(t *testing.T) {
	locksmith := &LockSmith{
		Node: &rpc.Node{
			Ring: make([]int, 0),
			RpcMap: make(map[int]chan *rpc.Data),
		},
		Nodes: make([]*rpc.Node, 0),
	}
	locksmith.InitializeNodes(3)
	if len(locksmith.Nodes) < 3 || len(locksmith.Node.Ring) < 3 || len(locksmith.Node.RpcMap) < 3 {
		t.Errorf("Expected 3 nodes to be created, but have incomplete creation: %d", len(locksmith.Nodes))
	}
	for _, node := range locksmith.Nodes {
		validate := validator.New()
		err := validate.Struct(node)
		if err != nil {
			t.Error(err)
		}
	}
}

// Expected HeartbeatTable to update to true when receive a heartbeat reply
func TestHandleMessageReceived(t *testing.T) {
	receivingChannel := make(chan *rpc.Data, 1)
	locksmith := &LockSmith{
		Node: &rpc.Node{
			RecvChannel: receivingChannel,
		},
		HeartBeatTable: make(map[int]bool),
	}
	locksmith.HeartBeatTable[1] = false
	go func(){
		time.Sleep(time.Second * 1)
		locksmith.Node.RecvChannel <- &rpc.Data{From: 1, Payload: map[string]interface{}{"type": "REPLY_HEARTBEAT"}}
	}()
	go func(){locksmith.HandleMessageReceived()
		if !locksmith.HeartBeatTable[1] {
			t.Errorf("Expected HeartbeatTable for Node 1 to be true, but instead it is still false.")
		}
	}()
}

// Expected 3 nodes to spin up and heartbeat table all update to true
func TestStartAllNodes(t *testing.T) {
	locksmith := &LockSmith{}
	locksmith.HeartBeatTable = make(map[int]bool)
	for i := 1; i <= 3; i ++ {
		newNode := &rpc.Node{
			Pid: i,
		}
		locksmith.Nodes = append(locksmith.Nodes, newNode)
	}
	locksmith.StartAllNodes()
	for pid, alive := range locksmith.HeartBeatTable {
		if !alive {
			t.Errorf("Expected Node [%d] to be alive, but yet it is not alive!", pid)
		}
	}
}

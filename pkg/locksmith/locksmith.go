package locksmith

import (
	"fmt"
	"time"

	"github.com/xmliszt/e-safe/config"
	"github.com/xmliszt/e-safe/pkg/rpc"
	"github.com/xmliszt/e-safe/util"
)

type LockSmith struct {
	LockSmithNode *rpc.Node `validate:"required"`
	Nodes []*rpc.Node `validate:"required"`
	HeartBeatTable map[int]bool `validate:"required"`
	DeadNode []int
}

// Start is the main function that starts the entire program
func Start() error {
	config, err := config.GetConfig()
	if err != nil {
		return err
	}

	locksmithServer := InitializeLocksmith()
	locksmithServer.InitializeNodes(config.Number)
	locksmithServer.StartAllNodes()
	go locksmithServer.DeadNodeChecker()

	fmt.Println("Locksmith [0] has started")
	go locksmithServer.CheckHeartbeat()	// Start periodically checking Node's heartbeat
	locksmithServer.HandleMessageReceived()	// Run this as the main go routine, so do not need to create separate go routine

	return nil
}

// InitializeLocksmith initializes the locksmith server object
func InitializeLocksmith() *LockSmith {
	receivingChannel := make(chan *rpc.Data, 1)
	sendingChannel := make(chan *rpc.Data, 1)
	isCoordinator := false
	locksmithServer := &LockSmith{
		LockSmithNode: &rpc.Node{
			IsCoordinator: &isCoordinator,
			Pid: 0,
			RecvChannel: receivingChannel,
			SendChannel: sendingChannel,
			Ring: make([]int, 0),
			RpcMap: make(map[int]chan *rpc.Data),
		},
		Nodes: make([]*rpc.Node, 0),
		HeartBeatTable: make(map[int]bool),
	}
	locksmithServer.LockSmithNode.RpcMap[0] = receivingChannel	// Add Locksmith receiving channel to RpcMap
	return locksmithServer
}

// InitializeNodes initializes the number n nodes that Locksmith is going to create
func (locksmith *LockSmith) InitializeNodes(n int) {
	for i := 1; i <= n; i ++ {
		nodeRecvChan := make(chan *rpc.Data, 1)
		nodeSendChan := make(chan *rpc.Data, 1)
		isCoordinator := false
		newNode := &rpc.Node{
			IsCoordinator: &isCoordinator,
			Pid: i,
			RecvChannel: nodeRecvChan,
			SendChannel: nodeSendChan,
		}
		locksmith.LockSmithNode.Ring = append(locksmith.LockSmithNode.Ring, i)
		locksmith.Nodes = append(locksmith.Nodes, newNode)
		locksmith.LockSmithNode.RpcMap[i] = nodeRecvChan
	}
	
	for _, node := range locksmith.Nodes {
		node.Ring = locksmith.LockSmithNode.Ring
		node.RpcMap = locksmith.LockSmithNode.RpcMap
	}
}

// HandleMessageReceived is a Go Routine to handle the messages received
func (locksmith *LockSmith) HandleMessageReceived() {
	for msg := range locksmith.LockSmithNode.RecvChannel {
		switch msg.Payload["type"] {
		case "REPLY_HEARTBEAT":
			locksmith.HeartBeatTable[msg.From] = true
		}
	}
}

// StartAllNodes starts up all created nodes
func (locksmith *LockSmith) StartAllNodes() {
	for _, node := range locksmith.Nodes {
		node.Start()
		locksmith.HeartBeatTable[node.Pid] = true
	}
}


// CheckHeartbeat periodically check if node is alive
func (locksmith *LockSmith) CheckHeartbeat() {
	config, err := config.GetConfig()
	if err != nil {
		fmt.Println("Fatal: Heartbeat checking has crashed. Reason: ", err)
		return
	}
	for {
		for _, pid := range locksmith.LockSmithNode.Ring {
			time.Sleep(time.Second * time.Duration(config.HeartbeatInterval))
			if locksmith.HeartBeatTable[pid] {
				go func(pid int) {
					locksmith.HeartBeatTable[pid] = false
					locksmith.LockSmithNode.SendSignal(pid, &rpc.Data{
						From: locksmith.LockSmithNode.Pid,
						To: pid,
						Payload: map[string]interface{}{
							"type": "CHECK_HEARTBEAT",
							"data": nil,
						},
					})
					time.Sleep(time.Second * 1)
					fmt.Println("Heartbeat Table: ", locksmith.HeartBeatTable)
					if !locksmith.HeartBeatTable[pid] {
						time.Sleep(time.Second * time.Duration(config.HeartBeatTimeout))
						if !locksmith.HeartBeatTable[pid] {
							fmt.Printf("Node [%d] is dead! Need to create a new node!\n", pid)
							found := util.IntInSlice(locksmith.DeadNode, pid)
							if !found {
								locksmith.DeadNode = append(locksmith.DeadNode, pid)
							}
							locksmith.SpawnNewNode(pid)
							time.Sleep(time.Second * time.Duration(config.NodeCreationTimeout))	// allow sufficient time for node to restart, then resume heartbeat checking
						}
					}
					}(pid)
				}
			}
		}
	}

// Check if a node is dead
func (locksmith *LockSmith) DeadNodeChecker() {
	config, err := config.GetConfig()
	if err != nil {
		fmt.Println("Fatal: Heartbeat checking has crashed. Reason: ", err)
		return
	}

	var potentialCandidate []int

	// Check what is the dead node, and assign the remaining alive node with the highest pid as coordinator
	if len(locksmith.DeadNode) != 0 {
		for idx := 1; idx <= config.Number; idx++ {
			found := util.IntInSlice(locksmith.DeadNode, idx)
				if !found {
					potentialCandidate = append(potentialCandidate, idx)
				}
				// TODO: spawn node here and remove from deadnode list
		}
	}
	coordinator := util.FindMax(potentialCandidate)
	locksmith.Election(coordinator)
	
}

// locksmith elects new node coordinator
func (locksmith *LockSmith) Election(pid int) {

}

// Spawn new nodes when a node is down
func (locksmith *LockSmith) SpawnNewNode(pid int) {
	
}


// TearDown terminates node, closes all channels
func (locksmith *LockSmith) TearDown() {
	close(locksmith.LockSmithNode.RecvChannel)
	close(locksmith.LockSmithNode.SendChannel)
	fmt.Printf("Locksmith Server [%d] has terminated!\n", locksmith.LockSmithNode.Pid)
}

// EndAllNodes starts teardown process of all created nodes
func (locksmit *LockSmith) EndAllNodes() {
	for _, node := range locksmit.Nodes {
		node.TearDown()
	}
}
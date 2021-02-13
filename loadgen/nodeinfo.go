package loadgen

import "github.com/google/uuid"

func newNodeInfo() nodeInfo {
	return nodeInfo{
		ID: uuid.New().String(),
	}
}

type nodeInfo struct {
	ID string
}

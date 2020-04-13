package root

import (
	"github.com/satori/go.uuid"
)

//Peer - micro service node for cluster
type Peer struct {
	ID      uuid.UUID
	Address string
}

//Cluster - of micro service
type Cluster struct {
    Members []*Peer
}

//Join cluster
func (c *Cluster) Join(p *Peer) {
    
}
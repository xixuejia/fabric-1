/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package solo

import (
	"time"

	ab "github.com/hyperledger/fabric/orderer/atomicbroadcast"
	"github.com/hyperledger/fabric/orderer/rawledger"
	"github.com/hyperledger/fabric/orderer/rawledger/ramledger"

	"github.com/op/go-logging"
	"google.golang.org/grpc"
)

var logger = logging.MustGetLogger("solo/server")

func init() {
	logging.SetLevel(logging.DEBUG, "")
}

// MagicLargestWindow is a temporary constant which limits the maximum window size, TODO this should be config at a later date
const MagicLargestWindow = 1000

type server struct {
	bs *broadcastServer
	ds *deliverServer
	rl rawledger.ReadWriter
}

// New creates a ab.AtomicBroadcastServer based on the solo orderer implementation
func New(queueSize, batchSize, historySize int, batchTimeout time.Duration, grpcServer *grpc.Server) ab.AtomicBroadcastServer {
	logger.Infof("Starting solo with queueSize=%d batchSize=%d historySize=%d batchTimeout=%v", queueSize, batchSize, historySize, batchTimeout)
	rl := ramledger.New(historySize)
	s := &server{
		bs: newBroadcastServer(queueSize, batchSize, batchTimeout, rl),
		ds: newDeliverServer(rl, MagicLargestWindow),
	}
	ab.RegisterAtomicBroadcastServer(grpcServer, s)
	return s
}

// Broadcast receives a stream of messages from a client for ordering
func (s *server) Broadcast(srv ab.AtomicBroadcast_BroadcastServer) error {
	return s.bs.handleBroadcast(srv)
}

// Deliver sends a stream of blocks to a client after ordering
func (s *server) Deliver(srv ab.AtomicBroadcast_DeliverServer) error {
	logger.Debugf("Starting new Deliver loop")
	return s.ds.handleDeliver(srv)
}

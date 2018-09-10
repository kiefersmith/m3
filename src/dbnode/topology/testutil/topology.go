// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package testutil

import (
	"fmt"

	"github.com/m3db/m3/src/dbnode/sharding"
	"github.com/m3db/m3/src/dbnode/topology"
	"github.com/m3db/m3cluster/shard"
)

const (
	SelfID = "self"
)

// MustNewTopologyMap returns a new topology.Map with provided parameters.
// It's a utility method to make tests easier to write.
func MustNewTopologyMap(
	replicas int,
	assignment map[string][]shard.Shard,
) topology.Map {
	v := NewTopologyView(replicas, assignment)
	m, err := v.Map()
	if err != nil {
		panic(err.Error())
	}
	return m
}

// NewTopologyView returns a new TopologyView with provided parameters.
// It's a utility method to make tests easier to write.
func NewTopologyView(
	replicas int,
	assignment map[string][]shard.Shard,
) TopologyView {
	total := 0
	for _, shards := range assignment {
		total += len(shards)
	}

	return TopologyView{
		HashFn:     sharding.DefaultHashFn(total / replicas),
		Replicas:   replicas,
		Assignment: assignment,
	}
}

// TopologyView represents a snaphshot view of a topology.Map.
type TopologyView struct {
	HashFn     sharding.HashFn
	Replicas   int
	Assignment map[string][]shard.Shard
}

// Map returns the topology.Map corresponding to a TopologyView.
func (v TopologyView) Map() (topology.Map, error) {
	var (
		hostShardSets []topology.HostShardSet
		allShards     []shard.Shard
		unique        = make(map[uint32]struct{})
	)

	for hostID, assignedShards := range v.Assignment {
		shardSet, _ := sharding.NewShardSet(assignedShards, v.HashFn)
		host := topology.NewHost(hostID, fmt.Sprintf("%s:9000", hostID))
		hostShardSet := topology.NewHostShardSet(host, shardSet)
		hostShardSets = append(hostShardSets, hostShardSet)
		for _, s := range assignedShards {
			if _, ok := unique[s.ID()]; !ok {
				unique[s.ID()] = struct{}{}
				uniqueShard := shard.NewShard(s.ID()).SetState(shard.Available)
				allShards = append(allShards, uniqueShard)
			}
		}
	}

	shardSet, err := sharding.NewShardSet(allShards, v.HashFn)
	if err != nil {
		return nil, err
	}

	opts := topology.NewStaticOptions().
		SetHostShardSets(hostShardSets).
		SetReplicas(v.Replicas).
		SetShardSet(shardSet)

	return topology.NewStaticMap(opts), nil
}

// SourceAvailableHost is a human-friendly way of constructing
// TopologyStates for test cases.
type SourceAvailableHost struct {
	Name        string
	Shards      []uint32
	ShardStates shard.State
}

// SourceAvailableHosts is a slice of SourceAvailableHost.
type SourceAvailableHosts []SourceAvailableHost

// TopologyState returns a topology.StateSnapshot that is equivalent to
// the slice of SourceAvailableHosts.
func (s SourceAvailableHosts) TopologyState(numMajorityReplicas int) *topology.StateSnapshot {
	topoState := &topology.StateSnapshot{
		Origin:           topology.NewHost(SelfID, "127.0.0.1"),
		MajorityReplicas: numMajorityReplicas,
		ShardStates:      make(map[topology.ShardID]map[topology.HostID]topology.HostShardState),
	}

	for _, host := range s {
		for _, shard := range host.Shards {
			hostShardStates, ok := topoState.ShardStates[topology.ShardID(shard)]
			if !ok {
				hostShardStates = make(map[topology.HostID]topology.HostShardState)
			}

			hostShardStates[topology.HostID(host.Name)] = topology.HostShardState{
				Host:       topology.NewHost(host.Name, host.Name+"address"),
				ShardState: host.ShardStates,
			}
			topoState.ShardStates[topology.ShardID(shard)] = hostShardStates
		}
	}

	return topoState
}

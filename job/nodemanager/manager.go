package nodemanager

import (
	"context"
	"fmt"
	"time"

	"github.com/yakumo-saki/phantasma-flow/global"
	"github.com/yakumo-saki/phantasma-flow/job/jobparser"
	"github.com/yakumo-saki/phantasma-flow/job/nodemanager/node"
	"github.com/yakumo-saki/phantasma-flow/util"
)

func (nm *NodeManager) GetCapacity(name string) int {
	const ERRVAL = -1
	log := util.GetLoggerWithSource(nm.GetName(), "GetCapacity")

	if nm.inShutdown {
		log.Warn().Msgf("NodeManager is in shutdown state. No jobs acceptable %s", name)
		return ERRVAL
	}

	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	pool, ok := nm.nodePool[name]

	if !ok {
		log.Warn().Msgf("No node registered for %s", name)
		return ERRVAL
	} else if pool.Len() == 0 {
		log.Warn().Msgf("No node is available for %s", name)
		return ERRVAL
	}

	meta := pool.Front().Value.(*nodeMeta)

	return meta.Capacity
}

// Exec job step.
// Logs and Results can be listen via messagehub
func (nm *NodeManager) ExecJobStep(ctx context.Context, step jobparser.ExecutableJobStep) {
	const NAME = "ExecJobStep"
	log := util.GetLoggerWithSource(nm.GetName(), NAME)
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// fetch first node from nodePool
	nd, ok := nm.nodePool[step.Node]
	if !ok {
		// XXX JOB FAIL #36
		log.Error().Msgf("Node '%s' is not found in NodeManager.", step.Node)
		return
	}

	// capacity check
	nodeMeta := nd.Front().Value.(*nodeMeta)
	nm.HasEnoughCapacity(nodeMeta, step) // check but not stop. TODO: return job start fail to caller

	// try initialize first
	execNode := node.ExecNode{}
	err := execNode.Initialize(nodeMeta.Def, step)
	if err != nil {
		// XXX job fail ? job hold ? #37
	}

	// Run on node
	nodeInst := nodeInstance{}
	nodeInst.UseCapacity = step.JobStepDefinition.UseCapacity
	ctx, cancel := context.WithCancel(ctx)
	nodeInst.Cancel = cancel // called on cleanUpNodePool

	nm.wg.Add(1)
	nodeMeta.Capacity = nodeMeta.Capacity - nodeInst.UseCapacity

	log.Trace().Msgf("node %s capa %v -> %v", step.Node, nodeMeta.Capacity+nodeInst.UseCapacity, nodeMeta.Capacity)

	execNode.Running = true // prevent cleanUp before goroutine start
	go execNode.Run(ctx, &nm.wg, step)

	// Register running node to nodemanager
	nodeMeta.RunningInstances[step.GetId()] = nodeInst
}

func (nm *NodeManager) HasEnoughCapacity(nodeMeta *nodeMeta, step jobparser.ExecutableJobStep) {
	log := util.GetLoggerWithSource(nm.GetName(), "HasEnoughCapacity")

	if nodeMeta.Capacity < step.JobStepDefinition.UseCapacity {
		msg := fmt.Sprintf("Insufficient node capacity node: %v, job-required: %v node_is_deprecated: %v",
			nodeMeta.Capacity, step.JobStepDefinition.UseCapacity, nodeMeta.Deprecated)
		if global.DEBUG {
			panic(msg)
		} else {
			log.Error().Msg(msg + ", continue anyway")
		}
	}
}

func (nm *NodeManager) DebugWaitForAllJobsDone() {
	log := util.GetLoggerWithSource(nm.GetName(), "WaitForAllJobsDone")

	count := 0
	for {
		nm.mutex.Lock()
		running := false
		for nodename, nodelist := range nm.nodePool {
			for e := nodelist.Front(); e != nil; e = e.Next() {
				nd := e.Value.(*nodeMeta)
				if len(nd.RunningInstances) > 0 {
					if count > 3 {
						log.Debug().Msgf("Still running on %s (%v)", nodename, len(nd.RunningInstances))
						count = 0
					} else {
						count++
					}
					running = true
					continue
				}
			}
		}

		nm.mutex.Unlock()
		if running {
			// wait for cleanup goroutine
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}

}

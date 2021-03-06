package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/yakumo-saki/phantasma-flow/global/consts"
	"github.com/yakumo-saki/phantasma-flow/job/jobparser"
	"github.com/yakumo-saki/phantasma-flow/messagehub"
	"github.com/yakumo-saki/phantasma-flow/pkg/message"
	"github.com/yakumo-saki/phantasma-flow/pkg/objects"
)

type ExecNode struct {
	nodeDef objects.NodeDefinition
	node    execNodeImpl
	Running bool
}

func (n *ExecNode) GetName() string {
	return "ExecNode"
}

func (n *ExecNode) Initialize(def objects.NodeDefinition, jobStep jobparser.ExecutableJobStep) error {
	n.nodeDef = def

	var impl execNodeImpl
	switch def.NodeType {

	case consts.NODE_TYPE_LOCAL:
		node := localExecNode{}
		impl = &node
	case consts.NODE_TYPE_SSH:
		node := sshExecNode{}
		impl = &node
	default:
		msg := fmt.Sprintf("Error Node %s: nodeType %s is unknown. %v", def.Id, def.NodeType, def)
		panic(msg)
	}

	err := impl.Initialize(def, jobStep)
	if err == nil {
		n.node = impl
	}
	return err
}

// Run jobStep context to cancel.
func (n *ExecNode) Run(ctx context.Context, wg *sync.WaitGroup, jobStep jobparser.ExecutableJobStep) {
	n.Running = true
	n.sendJobStepStartMsg(jobStep)
	n.sendJobStepStartLog(jobStep)

	exitcode := n.node.Run(ctx)

	success := true
	if exitcode < 0 {
		success = false // negative exitcode is always failure
	} else if exitcode > 0 {
		// TODO #43 Implement need exit code threshold
		success = false
	}

	n.Running = false
	n.sendJobStepEndMsg(jobStep, exitcode, success)
	n.sendJobStepEndLog(jobStep, exitcode, success)

	wg.Done()
}

func (n *ExecNode) sendJobStepStartMsg(jobstep jobparser.ExecutableJobStep) {
	msg := n.createExecuterMsg(jobstep, message.JOB_STEP_START)

	messagehub.Post(messagehub.TOPIC_JOB_REPORT, msg)
}

func (n *ExecNode) sendJobStepEndMsg(jobstep jobparser.ExecutableJobStep, exitcode int, success bool) {
	msg := n.createExecuterMsg(jobstep, message.JOB_STEP_END)
	msg.ExitCode = exitcode
	msg.Success = success
	messagehub.Post(messagehub.TOPIC_JOB_REPORT, msg)
}

func (n *ExecNode) createExecuterMsg(jobstep jobparser.ExecutableJobStep, subject string) *message.ExecuterMsg {
	msg := message.ExecuterMsg{}
	msg.Version = jobstep.Version
	msg.JobId = jobstep.JobId
	msg.RunId = jobstep.RunId
	msg.StepName = jobstep.Name
	msg.Node = jobstep.Node
	msg.Subject = subject
	// fmt.Printf("Job REPORT: %s, Job:%s/%s RunId:%s\n", subject, msg.JobId, msg.StepName, msg.RunId)

	return &msg
}

func (n *ExecNode) sendJobStepStartLog(jobstep jobparser.ExecutableJobStep) {
	lm := n.createJobLogMsg(jobstep)
	lm.Stage = objects.LM_STAGE_JOB
	lm.Message = fmt.Sprintf("Jobstep %s start.", jobstep.JobStepDefinition.Name)
	messagehub.Post(messagehub.TOPIC_JOB_LOG, lm)
}

func (n *ExecNode) sendJobStepEndLog(jobstep jobparser.ExecutableJobStep, exitcode int, success bool) {
	lm := n.createJobLogMsg(jobstep)
	lm.Stage = objects.LM_STAGE_JOB

	result := "SUCCESS"
	if !success {
		result = "FAILURE"
	}

	lm.Message = fmt.Sprintf("Jobstep %s ended. Result is %s. exitcode is %v.",
		jobstep.Name, result, exitcode)
	messagehub.Post(messagehub.TOPIC_JOB_LOG, lm)
}

func (n *ExecNode) createJobLogMsg(jobstep jobparser.ExecutableJobStep) *objects.JobLogMessage {
	lm := jobparser.CreateJobLogMsg(jobstep)
	lm.Source = n.GetName()
	lm.JobStep = jobstep.Name

	return lm
}

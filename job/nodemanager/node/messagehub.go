package node

import (
	"github.com/yakumo-saki/phantasma-flow/job/jobparser"
	"github.com/yakumo-saki/phantasma-flow/pkg/objects"
	"github.com/yakumo-saki/phantasma-flow/util"
)

func createJobLogMsg(seqNo uint64, jobStep jobparser.ExecutableJobStep) *objects.JobLogMessage {
	msg := objects.JobLogMessage{}
	msg.JobId = jobStep.JobId
	msg.RunId = jobStep.RunId
	msg.JobNumber = jobStep.JobNumber
	msg.Stage = objects.LM_STAGE_JOB
	msg.SeqNo = seqNo
	msg.Version = jobStep.Version
	msg.JobStep = jobStep.Name
	msg.Node = jobStep.Node
	msg.LogDateTime = util.GetDateTimeString()
	return &msg

}

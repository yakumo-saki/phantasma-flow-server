package message

import "github.com/yakumo-saki/phantasma-flow/pkg/objects"

// From repository to all

// reason
const JOB_STEP_START = "JOB-STEP-START" // reason: JOB STEP START
const JOB_STEP_END = "JOB-STEP-END"     // reason: JOB STEP END
const JOB_END = "JOB-END"               // reason: JOB END
const JOB_START = "JOB-START"           // reason: JOB START

type ExecuterMsg struct {
	// ExecuterMsg is message for Type = JOB_REPORT
	// This represents JOB or JOBSTEP start and end
	Subject  string // notification type. JOB_[STEP_]_(START|END)
	JobId    string
	RunId    string
	Version  objects.ObjectVersion
	StepName string // JOB_STEP_* only
	Node     string // JOB_STEP_* only, execute Node id
	Success  bool   // JOB_END or JOB_STEP_END
	Reason   string // why success = true / false
	ExitCode int    // JOB_STEP_END only
}

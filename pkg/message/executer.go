package message

// From repository to all

// reason
const JOB_STEP_START = "JOB-STEP-START" // reason: JOB STEP START
const JOB_STEP_END = "JOB-STEP-END"     // reason: JOB STEP END
const JOB_END = "JOB-END"               // reason: JOB END
const JOB_START = "JOB-START"           // reason: JOB START

type ExecuterMsg struct {
	Reason    string //
	JobId     string
	RunId     string
	StepName  string // JOB_STEP_* only
	JobResult string // JOB_END only
	ExitCode  int    // JOB_STEP_END only
}
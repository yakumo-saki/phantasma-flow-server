package objects

/*
kind: job-meta
jobId: job-id
meta:
  nextJobNumber: 3
results:
  - jobNumber: 1
    runId: gifLC7c1
    success: true
    version:
      major: 1
      minor: 0
    results:
      - stepname: step1
        success: true
      - stepname: step2
        success: true
  - jobNumber: 2
    success: false
	runId: IGTF5Z3i
    version:
      major: 1
      minor: 0
    results:
      - stepname: step1
        success: true
		exitCode: 0
      - stepname: step2
        success: false
		exitCode: 0
*/

type JobMetaLog struct {
	ObjectBase `json:",inline"` // "job-meta"
	Meta       JobMetaMeta      `json:"meta"`
	JobId      string           `json:"jobId"`
	Results    []*JobMetaResult `json:"results"`
}

type JobMetaMeta struct {
	NextJobNumber int `json:"nextJobNumber"`
}

// Result of 1 job execution
type JobMetaResult struct {
	JobNumber     int                  `json:"jobNumber"`
	Success       bool                 `json:"success"`
	Reason        string               `json:"reason"` // reason why job success is true or false
	Version       ObjectVersion        `json:"version"`
	RunId         string               `json:"runId"`
	StepResults   []*JobMetaStepResult `json:"stepResults"`
	StartDateTime string               `json:"startDateTime"` // RFC3306
	EndDateTime   string               `json:"endDateTime"`   // RFC3306
}

// Result of 1 job-step execution
type JobMetaStepResult struct {
	StepName      string `json:"stepName"`
	StartDateTime string `json:"startDateTime"` // RFC3306
	EndDateTime   string `json:"endDateTime"`   // RFC3306
	ExitCode      int    `json:"exitCode"`
	Success       bool   `json:"success"`
	Reason        string `json:"reason"` // reason why success is true / false
}

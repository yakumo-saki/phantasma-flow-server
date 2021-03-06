package executer

import (
	"fmt"
	"sync"

	"github.com/yakumo-saki/phantasma-flow/messagehub"
	"github.com/yakumo-saki/phantasma-flow/pkg/message"
	"github.com/yakumo-saki/phantasma-flow/pkg/objects"
	"github.com/yakumo-saki/phantasma-flow/util"
)

type execJobStepResult struct {
	Started bool // job step is started
	Ended   bool // job step is ended, regardless success or not
	Success bool // job step is success
}

// resultCollecter is
// * collecting JobStep result
// * Fail job, when job step failed
// * detect job is ended
func (ex *Executer) resultCollecter(startWg, stopWg *sync.WaitGroup) {
	const NAME = "resultCollecter"
	log := util.GetLoggerWithSource(ex.GetName(), NAME)
	log.Info().Msgf("Starting %s/%s.", ex.GetName(), NAME)

	jobEndCh := messagehub.Subscribe(messagehub.TOPIC_JOB_REPORT, ex.GetName())

	defer stopWg.Done()
	startWg.Done()

	for {
		select {
		case <-ex.RootCtx.Done():
			goto shutdown
		case msg, ok := <-jobEndCh:
			if !ok {
				continue
			}

			exeMsg := msg.Body.(*message.ExecuterMsg)

			switch exeMsg.Subject {
			case message.JOB_END:
				// job complete then delete from queue #39
				ex.mutex.Lock()
				delete(ex.jobQueue, exeMsg.RunId)
				ex.mutex.Unlock()
			case message.JOB_STEP_END:
				// log.Debug().Msgf("Got JOB_STEP_END %v", exeMsg)
				// step_end then store job result.
				// step_end then check return code and abort job if failed
				ex.mutex.Lock()
				qjob := ex.jobQueue[exeMsg.RunId]
				stepResult := qjob.StepResults[exeMsg.StepName]
				stepResult.Ended = true

				//
				if exeMsg.Success {
					// job step success. run next step by queueExecuter
					stepResult.Success = true

					reason := fmt.Sprintf("Job '%s' (runId:%s) , jobstep '%s' is success. exitcode is %v",
						exeMsg.JobId, exeMsg.RunId, exeMsg.StepName, exeMsg.ExitCode)
					log.Info().Msg(reason)
				} else {
					// job step failed. fail all jobsteps to prevent further run.
					stepResult.Success = false

					ex.failJobSteps(qjob, exeMsg.RunId)

					reason := fmt.Sprintf("Job '%s' (runId:%s) mark as failed, jobstep '%s' is failed. exitcode is %v",
						exeMsg.JobId, exeMsg.RunId, exeMsg.StepName, exeMsg.ExitCode)
					log.Info().Msg(reason)

					// send job end log
					qjob := ex.jobQueue[exeMsg.RunId]
					logmsg := ex.createJobLogMsg(qjob.Steps[0])
					logmsg.Stage = objects.LM_STAGE_POST
					logmsg.Message = reason
					messagehub.Post(messagehub.TOPIC_JOB_LOG, logmsg)

					// Send JOB_END message
					msg := ex.createExecuterMsg(qjob.Steps[0], message.JOB_END)
					msg.Success = false
					msg.Reason = reason
					messagehub.Post(messagehub.TOPIC_JOB_REPORT, msg)

					qjob.Cancel()
					goto exit
				}

				{ // check all jobstep is ended(whether success or not)
					end, success := ex.checkJobComplete(qjob)
					firstStep := qjob.Steps[0]
					if end {
						msg := ex.createExecuterMsg(firstStep, message.JOB_END)
						if success {
							msg.Success = true
							msg.Reason = "Job completed successfully"
						} else {
							msg.Success = false
							msg.Reason = "Job failed. Some jobstep is failed"
						}

						// send job end log
						qjob := ex.jobQueue[exeMsg.RunId]
						logmsg := ex.createJobLogMsg(firstStep)
						logmsg.Stage = objects.LM_STAGE_POST
						logmsg.Message = msg.Reason
						messagehub.Post(messagehub.TOPIC_JOB_LOG, logmsg)

						// send job end report
						messagehub.Post(messagehub.TOPIC_JOB_REPORT, msg)
						// log.Trace().Msgf("JOB_END_SENT %v", msg)

						qjob.Cancel()
					}
				}

			exit:
				ex.mutex.Unlock()
			default:
				continue
			}

		}
	}

shutdown:
	messagehub.Unsubscribe(messagehub.TOPIC_JOB_REPORT, ex.GetName())
	log.Debug().Msgf("%s/%s stopped.", ex.GetName(), NAME)
}

// checkJobComplete check all jobsteps are ended and all jobsteps are success
func (ex *Executer) checkJobComplete(qjob *queuedJob) (end, success bool) {
	end = true
	success = true

	for _, result := range qjob.StepResults {
		if !result.Ended {
			end = false
		}
		if !result.Success {
			success = false
		}
	}
	return end, success
}

func (ex *Executer) failJobSteps(qjob *queuedJob, runId string) {
	log := util.GetLoggerWithSource(ex.GetName(), "failJobSteps").With().
		Str("runId", runId).Logger()

	jobs := ex.jobQueue[runId]
	for step, result := range jobs.StepResults {
		if !result.Started && !result.Ended {
			result.Ended = true
			result.Success = false
			log.Debug().Msgf("Jobstep '%s' mark as failed, because of pre-jobstep is failed.", step)
		}
	}
}

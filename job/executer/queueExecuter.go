package executer

import (
	"sync"
	"time"

	"github.com/yakumo-saki/phantasma-flow/util"
)

// queueExecuter is find runnable jobstep from ex.jobQueue and exec it
func (ex *Executer) queueExecuter(startWg, stopWg *sync.WaitGroup) {
	const NAME = "queueExecuter"
	log := util.GetLoggerWithSource(ex.GetName(), NAME)
	log.Info().Msgf("Starting %s/%s.", ex.GetName(), NAME)

	defer stopWg.Done()
	startWg.Done()

	for {
		select {
		case <-ex.RootCtx.Done():
			goto shutdown
		case <-time.After(1 * time.Second):
			ex.mutex.Lock()
			for runId, queuedJob := range ex.jobQueue {
				ex.executeRunnable(runId, queuedJob)
			}
			ex.mutex.Unlock()
		}
	}

shutdown:
	// TODO need for cancel all jobs in running state. and wait for cancel done #38
	ex.mutex.Lock()
	for _, queuedJob := range ex.jobQueue {
		queuedJob.Cancel()
	}
	ex.mutex.Unlock()

	log.Debug().Msgf("%s/%s stopped.", ex.GetName(), NAME)
}

// executeRunnable find and run runnable job step from queuedJobs.
// Needs mutex.Lock before call.
func (ex *Executer) executeRunnable(runId string, job *queuedJob) {
	log := util.GetLoggerWithSource(ex.GetName(), "executeRunnable").
		With().Str("runId", runId).Logger()

	for _, step := range job.Steps {
		stat := job.StepResults[step.Name]

		if stat.Started || stat.Ended {
			// still running or already ended. nothing to do
			goto next
		}

		// not started and no presteps (= entrypoint)
		if len(step.PreSteps) == 0 {
			goto runIt
		}

		// check for all PreSteps are done and successful
		for _, pre := range step.PreSteps {
			s, ok := job.StepResults[pre]
			if !ok {
				log.Warn().Msgf("BUG? PreStep %s is not in StepResults", pre)
				goto next // no result = not started. not run. but this should not occur
			}
			if !s.Success {
				goto next // preStep is failed. not run (Job is failed.)
			} else {
				goto runIt
			}
		}

	runIt:
		log.Debug().Msgf("Jobstep start %s/%s", step.JobId, step.Name)
		ex.nodeMan.ExecJobStep(job.Context, step)
		stat.Started = true // prevent double exec, change started flag here (not wait for job_step_start msg)

	next:
		// next job step
	}

}

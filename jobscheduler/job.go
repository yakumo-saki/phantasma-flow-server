package jobscheduler

import (
	"time"

	"github.com/yakumo-saki/phantasma-flow/procman"
	"github.com/yakumo-saki/phantasma-flow/util"
)

type JobScheduler struct {
	procman.ProcmanModuleStruct
}

func (m *JobScheduler) IsInitialized() bool {
	return m.Initialized
}

func (m *JobScheduler) Initialize() error {
	m.Name = "JobScheduler"
	m.Initialized = true
	return nil
}

func (m *JobScheduler) GetName() string {
	return m.Name
}

func (js *JobScheduler) Start(procmanCh chan string) error {
	js.ProcmanCh = procmanCh
	log := util.GetLogger()

	log.Info().Msgf("Starting %s server.", js.GetName())
	js.ShutdownFlag = false

	for {
		select {
		case v := <-js.ProcmanCh:
			log.Debug().Msgf("Got request %s", v)
		default:
		}

		// todo Job Submitting

		if js.ShutdownFlag {
			break
		}

		time.Sleep(procman.MAIN_LOOP_WAIT)
	}

	log.Info().Msgf("%s Stopped.", js.GetName())
	js.ProcmanCh <- procman.RES_SHUTDOWN_DONE
	return nil
}

func (sv *JobScheduler) Shutdown() {
	// When shutdown initiated, procman calls this function.
	// All modules must send SHUTDOWN_DONE to procman before timeout.
	// Otherwise procman is not stop or force shutdown.

	log := util.GetLogger()
	log.Info().Msg("Shutdown initiated")
	sv.ShutdownFlag = true
}

func (js *JobScheduler) RequestHandler() {
	log := util.GetLogger()
	log.Debug().Msg("JobScheduler start")
}

package procmanExample

import (
	"time"

	"github.com/yakumo-saki/phantasma-flow/procman"
	"github.com/yakumo-saki/phantasma-flow/util"
)

type MinimalProcmanModule struct {
	procmanCh    chan string
	shutdownFlag bool
	Name         string // Recommended for debug
	initialized  bool
}

// returns this instance is initialized or not.
// When procman.Add, Procman calls Initialize() if not initialized.
func (m *MinimalProcmanModule) IsInitialized() bool {
	return m.initialized
}

// initialize this instance.
// Between Initialize and Start, no shutdown is called when error occures.
// so, dont initialize something needs shutdown sequence.
func (m *MinimalProcmanModule) Initialize() error {
	// used for procman <-> module communication
	// procman -> PAUSE(prepare for backup) is considered
	m.Name = "MinimalProcmanModule" // if you want to multiple instance, change name here
	m.initialized = true
	return nil
}

func (m *MinimalProcmanModule) GetName() string {
	// Name of module. must be unique.
	// return fix value indicates this module must be singleton.
	// add secondary instance of this module can cause panic by procman.Add
	return m.Name
}

// lets roll! Do not forget to save procmanCh from parameter.
func (m *MinimalProcmanModule) Start(procmanCh chan string) error {
	m.procmanCh = procmanCh
	log := util.GetLogger()

	log.Info().Msgf("Starting %s.", m.GetName())
	m.shutdownFlag = false

	go m.loop()

	// wait for other message from Procman
	for {
		select {
		case v := <-m.procmanCh:
			log.Debug().Msgf("Got request %s", v)
		default:
		}

		if m.shutdownFlag {
			break
		}

		time.Sleep(procman.MAIN_LOOP_WAIT) // Do not want to rush this loop
	}

	log.Info().Msgf("%s Stopped.", m.GetName())
	m.procmanCh <- procman.RES_SHUTDOWN_DONE
	return nil
}

func (m *MinimalProcmanModule) loop() {
	for {
		time.Sleep(procman.MAIN_LOOP_WAIT)
		if m.shutdownFlag {
			break
		}
	}

}

func (sv *MinimalProcmanModule) Shutdown() {
	// When shutdown initiated, procman calls this function.
	// All modules must send SHUTDOWN_DONE to procman before timeout.
	// Otherwise procman is not stop or force shutdown.

	log := util.GetLogger()
	log.Info().Msg("Shutdown initiated")
	sv.shutdownFlag = true
}

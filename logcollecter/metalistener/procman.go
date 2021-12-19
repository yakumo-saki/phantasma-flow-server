package metalistener

import (
	"context"
	"sync"
	"time"

	"github.com/yakumo-saki/phantasma-flow/procman"
	"github.com/yakumo-saki/phantasma-flow/util"
	"golang.org/x/sync/syncmap"
)

type MetaListener struct {
	procman.ProcmanModuleStruct
	logChannelsWg sync.WaitGroup
	logChannels   syncmap.Map // [string(runId)] <-chan LogMessage
	// logCloseFunc  syncmap.Map // [string(runId)] context.CancelFunc
}

func (m *MetaListener) IsInitialized() bool {
	return m.Initialized
}

func (m *MetaListener) Initialize() error {
	// used for procman <-> module communication
	// procman -> PAUSE(prepare for backup) is considered
	m.Initialized = true
	m.logChannels = syncmap.Map{}
	m.RootCtx, m.RootCancel = context.WithCancel(context.Background())
	return nil
}

func (m *MetaListener) GetName() string {
	return "MetaListener"
}

func (m *MetaListener) Start(inCh <-chan string, outCh chan<- string) error {
	m.FromProcmanCh = inCh
	m.ToProcmanCh = outCh
	log := util.GetLoggerWithSource(m.GetName(), "main")

	m.logChannelsWg = sync.WaitGroup{}
	m.logChannelsWg.Add(1)
	go m.LogMetaListener(m.RootCtx)

	log.Info().Msgf("Starting %s server.", m.GetName())

	time.Sleep(100 * time.Millisecond) // wait for LogMetaListener start

	m.ToProcmanCh <- procman.RES_STARTUP_DONE

	for {
		select {
		case v := <-m.FromProcmanCh:
			log.Debug().Msgf("Got request %s", v)
		case <-m.RootCtx.Done():
			goto shutdown
		}
	}

shutdown:
	m.logChannelsWg.Wait()
	log.Info().Msgf("%s Stopped.", m.GetName())
	m.ToProcmanCh <- procman.RES_SHUTDOWN_DONE
	return nil
}

func (sv *MetaListener) Shutdown() {
	log := util.GetLoggerWithSource(sv.GetName(), "shutdown")
	log.Debug().Msg("Shutdown initiated")
	sv.RootCancel()
}

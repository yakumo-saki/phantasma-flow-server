package messagehub_impl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
	"github.com/yakumo-saki/phantasma-flow/pkg/message"
	"github.com/yakumo-saki/phantasma-flow/util"
)

func (hub *MessageHub) Initialize() {
	hub.Name = "MessageHub"
	hub.listeners = sync.Map{}
	hub.listenerMutex = sync.Mutex{}
	hub.queue = goconcurrentqueue.NewFIFO()
	hub.senderWaitGroup = sync.WaitGroup{}
}

// XXX senderのctxはすべて記録しないとだめ。 unsubscribeできない。
// waitgroupは削除してもよさそうだが、確実に待つなら消せない
func (hub *MessageHub) StartSender() {
	if hub.senderCtx == nil {
		senderCtx, cancel := context.WithCancel(context.Background())
		hub.senderCtx = &senderCtx
		hub.senderCancel = &cancel
	}

	hub.senderWaitGroup.Add(1)
	go hub.Sender(hub.senderCtx)
}

// Stop sender thread. (Not waiting all queue done)
func (hub *MessageHub) StopSender() {
	log := util.GetLoggerWithSource(hub.Name, "stopSender")
	if hub.senderCtx == nil { // not start senders and shutdown
		log.Info().Msgf("StopSender: No senders started. Nothing to do.")
		return
	}

	log.Debug().Msgf("Wait for stopping all senders.")

	cancel := *hub.senderCancel
	cancel()

	hub.senderWaitGroup.Wait()
	log.Info().Msgf("Stopping all senders done.")
	hub.senderCancel = nil
	hub.senderCtx = nil
}

// Block new post and wait for queue become empty
func (hub *MessageHub) Shutdown() {
	log := util.GetLoggerWithSource(hub.Name, "shutdown")

	// Immediate shutdown, when called shutdown in sender stopped state
	if hub.senderCtx == nil {
		log.Warn().Int("queue_len", hub.queue.GetLen()).Msgf("Shutdown immediate. because of no sender started.")
		return
	}

	// context to timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// wait for queue flushed
	stop := false
	loopcount := 0
	for {

		select {
		case <-ctx.Done():
			log.Warn().Int("queue_len", hub.queue.GetLen()).Msgf("Shutdown timeout. Giving up send messages.")
			stop = true
		default:
			left := hub.queue.GetLen()
			stop = (left == 0)
			if loopcount > 9 {
				log.Info().Int("queue_len", hub.queue.GetLen()).Msgf("Shutdown in progress. Wait for all messages sent.")
				loopcount = 0
			}
		}

		if stop {
			goto shutdown
		}

		time.Sleep(500 * time.Millisecond)
		loopcount++
	}

shutdown:
	hub.StopSender()

	// dump if message left
	if hub.queue.GetLen() > 0 {
		for {
			m, err := hub.queue.Dequeue()
			if err == nil && m != nil {
				mx := m.(*message.Message)
				log.Error().Str("msg", fmt.Sprintf("%v", mx)).Msg("Dump messages can't send.")
			}
		}
	}

	log.Info().Msg("Shutdown done.")
}

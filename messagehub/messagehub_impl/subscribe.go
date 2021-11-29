package messagehub_impl

import (
	"github.com/yakumo-saki/phantasma-flow/pkg/message"
	"github.com/yakumo-saki/phantasma-flow/util"
)

// Accept new listener (=Subscriber)
func (hub *MessageHub) Subscribe(topic string, name string) chan *message.Message {
	log := util.GetLoggerWithSource(hub.Name, "listen")

	hub.listenerMutex.Lock()
	defer hub.listenerMutex.Unlock()

	arr, ok := hub.listeners.Load(topic)
	array := &[]listener{}
	if ok {
		array = arr.(*[]listener)
	}

	ch := make(chan *message.Message, 1)
	newListener := listener{name: name, ch: ch}
	ls := append(*array, newListener)

	hub.listeners.Store(topic, &ls)

	log.Debug().Str("name", name).Str("topic", topic).Int("listeners", len(ls)).Msgf("New listener added.")

	return ch
}

func (hub *MessageHub) Unsubscribe(topic string, name string) {
	log := util.GetLoggerWithSource(hub.Name, "stopListen").
		With().Str("topic", topic).Str("name", name).Logger()

	hub.listenerMutex.Lock()
	defer hub.listenerMutex.Unlock()

	arr, ok := hub.listeners.Load(topic)
	if !ok {
		// maybe not occured
		log.Error().Msg("Unsubscribe not listener, ignore")
		return
	}
	array := arr.(*[]listener)

	unsubed := false
	newListeners := make([]listener, len(*array))
	for _, lis := range *array {
		if lis.name != name {
			newListeners = append(newListeners, lis)
		} else {
			unsubed = true
			log.Debug().Msg("Unsubscribed")
			hub.senderWaitGroup.Done()
		}
	}
	hub.listeners.Store(topic, &newListeners)

	if !unsubed {
		log.Error().Msg("Listener not found, ignore")
	}
}
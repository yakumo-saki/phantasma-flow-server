package messagehub_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/huandu/go-assert"
	"github.com/yakumo-saki/phantasma-flow/messagehub"
	"github.com/yakumo-saki/phantasma-flow/messagehub/messagehub_impl"
)

func TestMessageHubSync(t *testing.T) {
	count := 0

	// null sending
	hub := messagehub_impl.MessageHub{}
	hub.Initialize()
	messagehub.SetMessageHub(&hub)

	msg := messagehub.NewMessage()
	msg.Topic = "topic1"
	msg.Body = "ABC1"
	hub.PostMsg(msg)
	count++

	// add listener
	wg := sync.WaitGroup{}
	wg.Add(3)
	go Listen(&hub, &wg, "topic1", "listner1")
	go Listen(&hub, &wg, "topic1", "listner2")
	go Listen(&hub, &wg, "topic1", "listner3")
	wg.Wait()

	hub.Post("topic1", "post test")
	count++

	fmt.Printf("messages sent %d\n", count)

	hub.Post("topic1", MSG_EXIT)

	hub.StartSender()
	hub.Shutdown()

	a := assert.New(t)
	a.Equal(count, getCount("listner1"))
	a.Equal(count, getCount("listner2"))
	a.Equal(count, getCount("listner3"))
}

package actor_client

import (
	"log"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	client, err := NewClient("localhost", "5432", "TRIvialPass")
	assert.Equal(t, err, nil, "Got Error")
	assert.NotEqual(t, client, nil, "Client should not be nil")
	list, err := client.List(-2)
	assert.Equal(t, err, nil, "Got Error")
	var some_actor_id int64
	log.Println("Listing")
	for actor_id, actor := range list {
		some_actor_id = actor_id
		log.Printf("Actor {%v}: %+v", actor_id, *actor)
	}
	log.Println("Done Listing")
	actor, err := client.Get(some_actor_id)
	assert.Equal(t, err, nil, "Got Error")
	log.Printf("Actor {%v}: %+v", actor.ActionId, *actor)
	actor, err = client.Get(math.MinInt64)
	assert.Equal(t, err, nil, "Got Error")
	log.Printf("Actor {%v}: %+v", actor.ActionId, *actor)
}

func TestWorkflow(t *testing.T) {
	// Connect
	client, err := NewClient("localhost", "5432", "TRIvialPass")
	assert.Equal(t, nil, err, "Got Error")
	assert.NotEqual(t, nil, client, "Client should not be nil")
	// First try getting running actor.
	running_actor, err := client.Get(-1)
	assert.Equal(t, nil, err, "Got Error")
	assert.Equal(t, running_actor.ActionId, int64(-1), "Got wrong actor")
	assert.Equal(t, int(running_actor.State), int(RUNNING), "Actor should be running")
	// Start
	actor, err := client.Start(-1, 0.25, 256.0, "MyArgsssss")
	if err != nil {
		log.Println(err)
	}
	assert.Equal(t, nil, err, "Got Error")
	log.Printf("Actor {%v}: %+v", actor.ActionId, *actor)
	var valid_state bool
	if actor.State == PENDING || actor.State == SCHEDULED {
		valid_state = true
	}
	assert.True(t, valid_state, "Invalid Start State")

	// Wait for running. Requires manual update to the DB
	log.Println("You have 30s to try manually giving an address to this actor.")
	retry := true
	for retry {
		actor, err = client.Get(actor.ActionId)
		assert.Equal(t, nil, err, "Got Error")
		if actor.State == RUNNING {
			retry = false
			log.Printf("Running Actor {%v}: %+v", actor.ActionId, *actor)
		} else if actor.State == PENDING || actor.State == SCHEDULED {
			time.Sleep(100 * time.Millisecond)
		} else {
			log.Print("Actor removed!")
			retry = false
		}
	}

	// Get
	actor2, err := client.Get(actor.ActionId)
	assert.Equal(t, nil, err, "Got Error")
	assert.Equal(t, actor2.ActionId, actor.ActionId, "Got wrong actor!!")
	valid_state = false
	if actor.State == PENDING || actor.State == SCHEDULED || actor.State == RUNNING {
		valid_state = true
	}
	assert.True(t, valid_state, "Invalid state!")

	// Stop
	running_state, err := client.Stop(actor.ActionId)
	assert.Equal(t, nil, err, "Got Error")
	assert.Equal(t, int(running_state), int(STOPPED), "Actor not stopped!")

	// Try getting again
	actor2, err = client.Get(actor.ActionId)
	assert.Equal(t, nil, err, "Got Error")
	valid_state = false
	if actor2.State == MISSING {
		valid_state = true
	}
	assert.True(t, valid_state, "Found stopped actor!!")
}

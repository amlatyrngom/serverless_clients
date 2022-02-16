package actor_client

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type ActorClient struct {
	db             *sql.DB
	leader_address string
	http_client    *http.Client
}

type Actor struct {
	ActionId int64
	State    RunningState
	Address  string
	Client   *ActorClient
	Conn     *http.Client
}

type ActorList map[int64]*Actor

func GetToJson(client *http.Client, url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func PostToJson(client *http.Client, url string, body []byte, target interface{}) (int, error) {
	r, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	defer r.Body.Close()
	code := r.StatusCode
	if code == http.StatusAccepted || code == http.StatusOK {
		return code, json.NewDecoder(r.Body).Decode(target)
	} else {
		return code, nil
	}
}

func (client *ActorClient) NewActor(action_id int64, running_state RunningState, address string) *Actor {
	return &Actor{
		ActionId: action_id,
		Client:   client,
		State:    running_state,
		Address:  address,
		Conn: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func NewClient(db_host string, db_port string, db_pwd string) (*ActorClient, error) {
	var client ActorClient
	// Connect to DB.
	conn_str := fmt.Sprintf("user=serverless dbname=serverless password=%s host=%s port=%s sslmode=disable", db_pwd, db_host, db_port)
	db, err := sql.Open("postgres", conn_str)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	client.db = db
	// Find actor manager
	err = client.FindLeader()
	if err != nil {
		log.Printf("Leader not yet found: %v\n", err)
		return &client, err
	}
	client.http_client = &http.Client{
		Timeout: time.Second * 10,
	}
	// Ping to test
	var resp PingResp
	err = GetToJson(client.http_client, fmt.Sprintf("http://%s/ping", client.leader_address), &resp)
	if err != nil {
		return &client, err
	}
	return &client, nil
}

func (client *ActorClient) FindLeader() error {
	// Lock actor_manager_lease table and read owner id, address.
	ADDRESS_QUERY := "SELECT owner_address FROM actor_manager_lease LIMIT 1"
	err := client.db.QueryRow(ADDRESS_QUERY).Scan(&client.leader_address)
	return err
}

func (client *ActorClient) List(deployment_id int64) (ActorList, error) {
	req := ListReq{
		DeploymentID: deployment_id,
	}
	var resp ListResp

	req_json, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, err = PostToJson(client.http_client, fmt.Sprintf("http://%s/list", client.leader_address), req_json, &resp)
	if err != nil {
		return nil, err
	}
	ret := make(ActorList)

	for action_id, status := range resp.States {
		ret[action_id] = client.NewActor(action_id, status.State, status.Address)
	}
	return ret, err
}

func (client *ActorClient) Start(deployment_id int64, cpus float32, mem float32, args string) (*Actor, error) {
	req := StartReq{
		DeploymentId: deployment_id,
		Cpus:         cpus,
		Mem:          mem,
		Args:         args,
	}
	req_json, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var resp StartResp
	_, err = PostToJson(client.http_client, fmt.Sprintf("http://%s/start_action", client.leader_address), req_json, &resp)
	if err != nil {
		return nil, err
	}
	return client.NewActor(resp.ActionID, resp.State, ""), nil
}

func (client *ActorClient) Get(actor_id int64) (*Actor, error) {
	req := StatusReq{
		ActionId: actor_id,
	}
	req_json, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var resp StatusResp
	_, err = PostToJson(client.http_client, fmt.Sprintf("http://%s/action_status", client.leader_address), req_json, &resp)
	if err != nil {
		return nil, err
	}
	return client.NewActor(actor_id, resp.State, resp.Address), nil
}

func (client *ActorClient) Stop(actor_id int64) (RunningState, error) {
	req := StopReq{
		ActionId: actor_id,
	}
	req_json, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	var resp StopResp
	_, err = PostToJson(client.http_client, fmt.Sprintf("http://%s/stop_action", client.leader_address), req_json, &resp)
	if err != nil {
		return 0, err
	}
	return resp.State, nil
}

func (client *ActorClient) FindDeploymentID(name string) (int64, error) {
	Q := "SELECT id FROM deployments WHERE name=$1"
	var id int64
	err := client.db.QueryRow(Q, name).Scan(&id)
	return id, err
}

func (actor *Actor) HttpURL(route string) string {
	return fmt.Sprintf("http://%s/%s", actor.Address, route)
}

// Untested
func (actor *Actor) WaitForRunning(res_ch chan bool, done_ch chan bool) {
	wait_ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-wait_ticker.C:
			new_actor, err := actor.Client.Get(actor.ActionId)
			if err != nil {
				log.Printf("ActorClient error %v", err)
				res_ch <- false
				return
			}
			if new_actor.State == RUNNING {
				*actor = *new_actor
				res_ch <- true
				return
			}
		case <-done_ch:
			return
		}
	}
}

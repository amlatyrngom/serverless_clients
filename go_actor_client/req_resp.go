package actor_client

type RunningState int

const (
	RUNNING       RunningState = iota
	SCHEDULED                  = iota
	PENDING                    = iota
	UNSATISFIABLE              = iota
	NOTLEADER                  = iota
	STOPPED                    = iota
	NOTDEPLOYED                = iota
	MISSING                    = iota
)

type SchedulingMechanism int

const (
	DEFAULT_SCHEDULE SchedulingMechanism = iota
	ANTI_AFFINITY                        = iota
)

type StartReq struct {
	DeploymentId int64   `json:"deployment_id"`
	Cpus         float32 `json:"cpus"`
	Mem          float32 `json:"mem"`
	Args         string  `json:"args"`
	// TODO: Add scheduling mechanism
}

type StopReq struct {
	ActionId int64 `json:"action_id"`
}

type StatusReq struct {
	ActionId int64 `json:"action_id"`
}

type ListReq struct {
	DeploymentID int64 `json:"deployment_id"`
}

type StartResp struct {
	ActionID int64        `json:"action_id"`
	State    RunningState `json:"state"`
}

type StopResp struct {
	State RunningState `json:"state"`
}

type StatusResp struct {
	State   RunningState `json:"state"`
	Address string       `json:"address"`
	// TODO: add cpu, mem, etc.
}

type ListResp struct {
	States map[int64]StatusResp `json:"states"`
}

type PingResp struct {
	Message string `json:"message"`
}

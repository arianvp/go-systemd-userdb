package userdb

type UserFields struct {
	UserName string  `json:"userName"`
	Uid      *uint32 `json:"uid,omitempty"`
	Gid      *uint32 `json:"gid,omitempty"`
}

type PerMachine struct {
	MatchMachineId string `json:"matchMachineId,omitempty"`
	MatchHostname  string `json:"matchHostname,omitempty"`
	UserFields
}

type Status struct {
	DiskUsage   *uint64 `json:"diskUsage,omitempty"`
	DiskFree    *uint64 `json:"diskFree,omitempty"`
	DiskSize    *uint64 `json:"diskSize,omitempty"`
	DiskCeiling *uint64 `json:"diskCeiling,omitempty"`
	DiskFloor   *uint64 `json:"diskFloor,omitempty"`
	State       string  `json:"state,omitempty"`
	Service     string  `json:"service,omitempty"`
}

type UserRecord struct {
	UserFields
	Privileged *UserFields           `json:"privileged,omitempty"`
	Binding    map[string]UserFields `json:"binding,omitempty"`
	PerMachine *PerMachine           `json:"perMachine,omitempty"`
	Status     map[string]Status     `json:"status,omitempty"`
}

type GetUserRecordRequestParams struct {
	UserName *string `json:"userName,omitempty"`
	Uid      *uint32 `json:"uid,omitempty"`
	Service  string  `json:"service"`
}

type GetUserRecordRequest struct {
	Method     string                     `json:"method"`
	Parameters GetUserRecordRequestParams `json:"parameters"`
	More       bool                       `json:"more"`
}

type GetUserRecordReplyParams struct {
	Record UserRecord `json:"record"`
}

type GetUserRecordReply struct {
	Parameters GetUserRecordReplyParams `json:"parameters"`
	Continues  bool                     `json:"continues,omitempty"`
	Error      string                   `json:"error,omitempty"`
}

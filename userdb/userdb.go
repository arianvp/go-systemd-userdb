package userdb

type UserRecord struct {
	UserName string `json:"userName"`
	// TODO other fields
}

type GetUserRecordRequestParams struct {
	UserName *string `json:"userName,omitempty"`
	Uid      *int64  `json:"uid,omitempty"`
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

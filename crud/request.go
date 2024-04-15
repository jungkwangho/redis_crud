package main

type CheckRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CheckResult struct {
	RemoteError  string `json:"remote_error"`
	RemoteResult string `json:"remote_result"`
}

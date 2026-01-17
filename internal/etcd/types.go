package etcd

import clientv3 "go.etcd.io/etcd/client/v3"

type KeyValue struct {
	Key          string
	Value        string
	ValuePreview string
}

type ConnectionMsg struct {
	Client  *clientv3.Client
	Success bool
	Err     error
}

type KeysMsg struct {
	Keys    []KeyValue
	HasMore bool
	Err     error
}

type ValueMsg struct {
	Key   string
	Value string
	Err   error
}

type CountMsg struct {
	Count int
	Err   error
}

package db

type Instance any

type Connector interface {
	Connect() *Instance
}

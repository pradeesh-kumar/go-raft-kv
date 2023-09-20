package kvclient

type SessionManager interface {
	GetSession() Session
}

type DefaultKvSessionManager struct {
}

type Session struct {
	clientId string
}

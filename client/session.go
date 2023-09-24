package kvclient

type SessionManager interface {
	GetSession() *Session
}

type DefaultSessionManager struct {
}

type Session struct {
	clientId string
}

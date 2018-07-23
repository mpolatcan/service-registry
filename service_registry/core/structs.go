package core

type Service struct {
	ServiceName string
	ServiceHostname string
	ServicePort int
	ServiceHeartbeatEndpoint string
}

type Observer struct {
	ObservedServices string
	ObserverHostname string
	ObserverPort int
	ObserverUpdateEndpoint string
}
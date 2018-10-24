package main

import "sync"

type (
	awake struct {
		timeout  chan struct{}
		redirect chan struct{}
	}
	awakingApps struct {
		mutex *sync.Mutex
		state map[string]awake
	}
)

func newAwakingApps() *awakingApps {
	return &awakingApps{
		state: make(map[string]awake),
		mutex: &sync.Mutex{},
	}
}

func (await awakingApps) registerApp(app string) bool {
	await.mutex.Lock()
	defer await.mutex.Unlock()
	if _, ok := await.state[app]; ok {
		return false
	}
	await.state[app] = newAwake()
	return true
}

func newAwake() awake {
	timeout := make(chan struct{})
	redirect := make(chan struct{})
	return awake{timeout, redirect}
}

func (await awakingApps) delete(app string) {
	await.mutex.Lock()
	delete(await.state, app)
	await.mutex.Unlock()
}

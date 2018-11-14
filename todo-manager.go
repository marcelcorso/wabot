package main

import (
	"sync"
)

type todoManager struct {
	sync.RWMutex
	m map[string]*todoList
}

func newTodoManager() *todoManager {
	return &todoManager{
		m: make(map[string]*todoList),
	}
}

func (tm *todoManager) fetch(key string) *todoList {
	tm.Lock()
	todo, ok := tm.m[key]
	if !ok {
		todo = newTodoList()
		tm.m[key] = todo
	}
	tm.Unlock()

	return todo
}

func (tm *todoManager) close(key string) {
	tm.Lock()
	delete(tm.m, key)
	tm.Unlock()
}

package main

import (
	"errors"
	"sync"
)

type todoList struct {
	sync.RWMutex
	list []string
}

func newTodoList() *todoList {
	return &todoList{
		list: []string{},
	}
}

var errItemNotFound = errors.New("item not found")

func (l *todoList) done(i int) error {
	if (i > len(l.list)-1) || (i < 0) {
		return errItemNotFound
	}
	l.Lock()
	// just delete
	l.list = append(l.list[:i], l.list[i+1:]...)
	l.Unlock()
	return nil
}

func (l *todoList) add(key string) {
	l.Lock()
	l.list = append(l.list, key)
	l.Unlock()
}

func (l *todoList) read() []string {
	return l.list
}

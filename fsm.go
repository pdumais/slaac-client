package main

import (
    "time"
)

type TimerFunc func(s *State)
type AdvFunc func(s *State, msg AdvMessage) 
type EnterFunc func(s *State)

type Handlers struct {
    onEnter EnterFunc
    onTimer TimerFunc
    onAdvertisement AdvFunc
}

type State struct {
    currentHandlers *Handlers
    ndp *NDP
    timer *time.Ticker
}


func NewFSM() (*State){
    var s State

    s.currentHandlers = nil

    return &s
}

func TransitionTo(s *State, h *Handlers){
    if (s.currentHandlers == h) {
        return
    }
    s.currentHandlers = h
    if (h.onEnter != nil) {
        h.onEnter(s)
    }
    
}

package main

import (
    "fmt"
    "time"
    "github.com/mdlayher/ndp"
)

var initialState Handlers
var solicitingState Handlers
var configuredState Handlers
var fsm *State

func AutoConfig(iface string) {
    //TODO: check in /proc if RA is handled by kernel. This would create conflicts
 
    initialState.onEnter = onEnterInitial  
    solicitingState.onTimer = onTimerSoliciting
    solicitingState.onAdvertisement = onAdvertisementSoliciting
    solicitingState.onEnter = onEnterSoliciting
    configuredState.onEnter = onEnterConfigured
    configuredState.onAdvertisement = onAdvertisementConfigured

    fsm = NewFSM()
    fsm.ndp = NewNDP(iface)

    TransitionTo(fsm, &initialState)

    for {
        select {
            case msg := <- fsm.ndp.msgChannel:
                if (fsm.currentHandlers.onAdvertisement != nil) {
                   fsm.currentHandlers.onAdvertisement(fsm, msg)
                }
        }
    }
}


func onEnterInitial(s *State) {
    fmt.Println("Initial")
    TransitionTo(fsm, &solicitingState)
}

func onEnterSoliciting(s *State) {
    s.timer = time.NewTicker(3*time.Second)

    go func() {
        for _ = range s.timer.C {
            fsm.currentHandlers.onTimer(fsm)
        }
    }()

    Solicit(s.ndp)
}

func onEnterConfigured(s *State) {
    s.timer.Stop()
}

func onTimerSoliciting(s *State) {
    fmt.Println("On Timer")
    Solicit(s.ndp)
}


// Once the interface is configured, we do not accept anymore router advertisements.
func onAdvertisementConfigured(s *State, msg AdvMessage) {
    fmt.Println("Ignoring router advertisement since we are already configured")
}

func onAdvertisementSoliciting(s *State, msg AdvMessage) {
    fmt.Println("Processing Router Advertisement")

    ifconfig := InterfaceConfiguration{}
    ifconfig.Managed = msg.msg.ManagedConfiguration
    ifconfig.OtherConfiguration = msg.msg.OtherConfiguration
    ifconfig.Gateway = msg.from

    for _, o := range msg.msg.Options {
        switch o :=  o.(type) {
            case *ndp.PrefixInformation:
                //TODO: should handle lifetime parameter
                ifconfig.Prefix = o.Prefix
                ifconfig.PrefixLength = o.PrefixLength
            case *ndp.RecursiveDNSServer:
                //TODO: should handle lifetime parameter
                ifconfig.DNSServers = o.Servers
        }
    }

    err := ConfigureInterface(s.ndp.dev, &ifconfig)
    if (err != nil) {
        fmt.Println("Error processing router advertisement: ", err)
        return
    }
    
    TransitionTo(fsm, &configuredState)

}

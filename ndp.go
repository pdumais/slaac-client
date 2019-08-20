package main

import (
    "fmt"
    "net"
    "github.com/mdlayher/ndp"
)

type AdvMessage struct {
    msg *ndp.RouterAdvertisement
    from net.IP
}

type NDP struct{
    iface string
    connection *ndp.Conn
    dev *net.Interface
    msgChannel chan AdvMessage
}

func NewNDP(iface string) (*NDP) {
    var n NDP

    n.iface = iface

    dev, err := net.InterfaceByName(iface)
    if err != nil {
        fmt.Println("failed to get interface: ",err)
        return nil
    }

    c, ip, err := ndp.Dial(dev, ndp.LinkLocal)
    if err != nil {
        fmt.Println("failed to dial NDP connection: ", err)
        return nil
    }
    fmt.Println("ndp: bound to address:", ip)

    n.dev = dev
    n.connection = c


    ch := make(chan AdvMessage)
    go func() {
        for {
            msg, _, from, err := c.ReadFrom()
            if err != nil {
                fmt.Println("failed to read NDP message: ", err)
            }

            ra, ok := msg.(*ndp.RouterAdvertisement)
            fmt.Println("Received Router Advertisement")
            if (ok) {
                ch <- AdvMessage{ra, from}
            }
        }
    }()
    n.msgChannel = ch

    return &n
}

func Solicit(n *NDP) {

    m := &ndp.RouterSolicitation{
        Options: []ndp.Option{
            &ndp.LinkLayerAddress{
               Direction: ndp.Source,
               Addr:      n.dev.HardwareAddr,
           },
       },
    }

    fmt.Println("Sending Router Solicitation on "+n.iface)
    if err := n.connection.WriteTo(m, nil, net.IPv6linklocalallrouters); err != nil {
        fmt.Println("failed to write router solicitation: ", err)
    }

}

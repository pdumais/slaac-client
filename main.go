package main

import (
    "fmt"
    "flag"
)


func main() {

    iface := flag.String("i", "eth1", "Network Interface")
    flag.Parse()

    fmt.Println("Interface "+*iface+" chosen");
    AutoConfig(*iface)

}

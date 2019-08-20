package main

import (
    "fmt"
    "net"
    "strconv"
    "os/exec"
    "github.com/mdlayher/eui64"
)

type InterfaceConfiguration struct {
    Managed bool
    OtherConfiguration bool
    Prefix net.IP
    PrefixLength uint8
    Gateway net.IP
    DNSServers  []net.IP
}

func ConfigureInterface(dev *net.Interface, config *InterfaceConfiguration) (error){

    err := setGateway(dev, config.Gateway.String()) 
    if (err != nil) {
        return err
    }

    if (config.Managed) {
        fmt.Println("This interface will be managed by DHCPv6")
        //TODO: we must retain handle of this command in case we wanna kill it
        cmd := exec.Command("dhclient", "-6", "-nw", "-1", "-q", "-lf", "/var/lib/dhclient/dhclient--"+dev.Name+".lease", "-pf", "/var/run/dhclient-"+dev.Name+".pid", dev.Name)
        err := cmd.Start()
        return err
    }

    if (config.Prefix == nil) {
        return fmt.Errorf("No prefix information found")
    }

    eui64Ip, err := eui64.ParseMAC(config.Prefix, dev.HardwareAddr)
    if (err != nil) {
        return fmt.Errorf("Generating eui64 IP: %v",err)
    }

    err = setIP(dev, eui64Ip.String(), config.PrefixLength)
    if (err != nil) {
        return err
    }

    //TODO: add support for privacy extension

    if (config.DNSServers != nil) {
        fmt.Println("DNS Servers were included in this advertisement")
        setDNS(config.DNSServers)
    }


    if (config.OtherConfiguration) {
        fmt.Println("This interface will obtain more configuration from  DHCPv6")
        cmd := exec.Command("dhclient", "-6", "-S", dev.Name)
        err := cmd.Run()
        return err
    }

    return nil
}

func setIP(dev *net.Interface, ip string, pl uint8) (error){
    fmt.Println("IP will be: ",ip)

    ipl := int(128-pl)
    ip = ip+"/"+strconv.Itoa(ipl)
    fmt.Println(ip)
    cmd := exec.Command("ip","add","add",ip,"dev",dev.Name)
    err := cmd.Run()

    return err
} 

func setGateway(dev *net.Interface, ip string) (error){
    fmt.Println("Gateway will be: ",ip)

    cmd := exec.Command("ip","route","add","default","via",ip,"dev",dev.Name)
    err := cmd.Run()

    return err
}

func setDNS(servers []net.IP) (error){
    
    var lines string
    for _, v := range servers {
        lines += "nameserver "+v.String()+" # slaac \n"
    }

    fmt.Println(lines)
    cmd := exec.Command("bash", "-c", "echo -e '"+lines+"' > /etc/resolv.conf")
    err := cmd.Run()
    return err
}

package main

import (
	"flag"
	"fmt"
	"github.com/segator/wireguard-dynamic/mesh"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type cmdLineOpts struct {
	init bool
	join bool
	mesh mesh.Mesh
	meshPeer mesh.MeshLocalPeer
}

var (
	opts cmdLineOpts
	initCmd   = flag.NewFlagSet("init", flag.ExitOnError)
	joinCmd   = flag.NewFlagSet("join", flag.ExitOnError)
)

func main() {
	chanOSSignal := make(chan os.Signal)
	signal.Notify(chanOSSignal, os.Interrupt, syscall.SIGTERM)
	subnets := ""
	privateIps := ""
	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		opts.init=true
	case "join":
		opts.join=true
		joinCmd.StringVar(&opts.mesh.MeshID, "token", "", "join  mesh token")
		joinCmd.IntVar(&opts.meshPeer.ListenPort, "listen-port", 31111, "Listen Port for VPN Service")
		joinCmd.IntVar(&opts.meshPeer.ApiListenPort, "api-listen-port", 23103, "Listen Port for REST API Service")
		joinCmd.IntVar(&opts.meshPeer.PublicPort, "public-port", 31111, "Public listen Port for the VPN Service, if connection through NAT can be diferent than listen port")
		joinCmd.StringVar(&opts.meshPeer.PublicIP, "public-ip", "auto", "Public IP used by other nodes to connect to this node, by default is auto calculated")
		joinCmd.StringVar(&privateIps, "private-ip", "auto", "Private IP used by other nodes on the same LAN to connect to this node directly without tunneling through wireguard, you can also define ranges so we will search for privateIps that match this range(Example: 192.168.0.0/16), by default will use all Ip's on internal net device,if you don't want this node be reached by internal network set 'none'")

		joinCmd.StringVar(&opts.meshPeer.VPNIP, "vpn-ip", "auto", "IP of the internal VPN this node will have, only /32 allowed")
		joinCmd.StringVar(&subnets, "accept-networks", "", "network list splited with ,(coma) so other Nodes will know how to achieve those subnets through this node")
		//joinCmd.StringVar(&subnets, "accept-networks-routing", "NONE", "how other nodes will route to accepted networks of this node(MASQUERADE, NONE, FORWARD)")

		joinCmd.IntVar(&opts.meshPeer.KeepAlive, "keep-alive", 15, "Keep Alive in seconds")
		joinCmd.StringVar(&opts.meshPeer.DeviceName, "device-name", "wg0", "Device name, by default wg0")
		joinCmd.Parse(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "wireguard-dynamic init #this create new mesh token\nwireguard-dynamic join #this join this node to a existing mesh")
		os.Exit(1)
	}

	//Get Repo Type
	storeRepository := mesh.NewKVDBRepository()
	//Get VPN Type( Chaining Wireguard type + host Gateway)
	networkService :=  mesh.NewHostGatewayNetworkService(mesh.NewWireGuardNetworkService())

	//Mesh Type
	me :=mesh.NewSimpleMeshService(storeRepository,networkService)
	//Api Service
	restApiService := mesh.NewRestAPIService(opts.meshPeer.ApiListenPort,me)
	if opts.init {
		mesh :=me.CreateMesh()
		log.Println(mesh.MeshID)
	}else if opts.join{
		opts.meshPeer.Version=0
		if opts.meshPeer.PublicIP == "auto" {
			opts.meshPeer.AutoPublicIP=true
			opts.meshPeer.PublicIP = ""
		}
		if privateIps == "auto" {
			opts.meshPeer.PrivateIPs = findAllPrivateIps()
		}else if privateIps == "none" {
			opts.meshPeer.PrivateIPs = []string{"none"}
		}else{
			allPrivateIPs:=findAllPrivateIps()
			opts.meshPeer.PrivateIPs = []string{}
			privateIpsArray :=strings.Split(privateIps,",")
			for _, privateIP := range privateIpsArray {
				ip, subnet,err :=net.ParseCIDR(privateIP)
				if err != nil {
					opts.meshPeer.PrivateIPs = append(opts.meshPeer.PrivateIPs,privateIP)
				}else{
					log.Println(ip.String() + " " + subnet.String())
					for _, privateLocalIP := range allPrivateIPs {
						if subnet.Contains(net.ParseIP(privateLocalIP)) {
							addIP:=true
							for _, alreadyAddedIP := range opts.meshPeer.PrivateIPs {
								if alreadyAddedIP == privateLocalIP {
									addIP=false
									break
								}
							}
							if addIP{
								opts.meshPeer.PrivateIPs = append(opts.meshPeer.PrivateIPs,privateLocalIP)
							}
						}
					}
				}
			}
		}

		if opts.meshPeer.VPNIP == "auto" {
			opts.meshPeer.AutoVPNIP=true
			opts.meshPeer.VPNIP = ""
		}
		if subnets != "" {
			opts.meshPeer.AllowedIPs = strings.Split(subnets,",")
		}
		go restApiService.Listen()
		me.JoinMesh(opts.mesh,opts.meshPeer)
		<-chanOSSignal
		log.Println("Signal detected, closing network")
		me.Stop()
	}

	os.Exit(0)
}

func findAllPrivateIps() []string {
	foundIPS :=  []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Panic("Oops:"+err.Error())
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				foundIPS = append(foundIPS,ipnet.IP.String())
			}
		}
	}
	return foundIPS
}

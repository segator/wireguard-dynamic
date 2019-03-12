package main

import (
	"flag"
	"fmt"
	"github.com/segator/wireguard-dynamic/mesh"
	"log"
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
	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		opts.init=true
	case "join":
		opts.join=true
		joinCmd.StringVar(&opts.mesh.MeshID, "token", "", "join  mesh token")
		joinCmd.IntVar(&opts.meshPeer.ListenPort, "listen-port", 31111, "Listen Port for the VPN Service")
		joinCmd.IntVar(&opts.meshPeer.PublicPort, "public-port", 31111, "Public listen Port for the VPN Service, if connection through NAT can be diferent than listen port")
		joinCmd.StringVar(&opts.meshPeer.PublicIP, "public-ip", "auto", "Public IP used by other nodes to connect to this node, by default is auto calculated")
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
	//Get VPN Type
	networkService :=mesh.NewWireGuardNetworkService()

	//Mesh Type
	me :=mesh.NewSimpleMeshService(storeRepository,networkService)
	if opts.init {
		mesh :=me.CreateMesh()
		log.Println(mesh.MeshID)
	}else if opts.join{
		opts.meshPeer.Version=0
		if opts.meshPeer.PublicIP == "auto" {
			opts.meshPeer.AutoPublicIP=true
			opts.meshPeer.PublicIP = ""
		}

		if opts.meshPeer.VPNIP == "auto" {
			opts.meshPeer.AutoVPNIP=true
			opts.meshPeer.VPNIP = ""
		}
		if subnets != "" {
			opts.meshPeer.AllowedIPs = strings.Split(subnets,",")
		}
		me.JoinMesh(opts.mesh,opts.meshPeer)
		<-chanOSSignal
		log.Println("Signal detected, closing network")
		me.Stop()
	}

	os.Exit(0)
}

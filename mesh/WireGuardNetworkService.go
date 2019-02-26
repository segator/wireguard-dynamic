package mesh

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"github.com/segator/wireguard-dynamic/cmd"
)
type WireGuardNetworkService struct {
	InterfaceName string
	InterfaceMTU int
}


func NewWireGuardNetworkService() NetworkService {
	return &WireGuardNetworkService{
		InterfaceName:"wg0",
		InterfaceMTU: 1420,
	}
}

func (wireguard *WireGuardNetworkService) InitializeNetworkDevice(peer *MeshLocalPeer) {
	//Delete if exists
	cmd.Command("ip","link","delete","dev",wireguard.InterfaceName)
	cmd.Command("iptables","-D","FORWARD","-i",wireguard.InterfaceName ,"-j","ACCEPT")

	fmt.Println("---->Initializing Wireguard Device " + wireguard.InterfaceName)


    exitCode  := cmd.Command("ip","link","add",wireguard.InterfaceName,"type","wireguard")
    cmd.ValidateCommand(exitCode)

	exitCode  = cmd.Command("ip","address","add",peer.VPNIP+"/32","dev",wireguard.InterfaceName)
	cmd.ValidateCommand(exitCode)

	exitCode = cmd.Command("ip","link","set","mtu",strconv.Itoa(wireguard.InterfaceMTU),"up","dev",wireguard.InterfaceName)
	cmd.ValidateCommand(exitCode)

	exitCode  = cmd.Command("iptables","-A","FORWARD","-i",wireguard.InterfaceName ,"-j","ACCEPT")
	cmd.ValidateCommand(exitCode)

	//create keys
	privateKey,err :=cmd.CommandCaptureOutput("wg","genkey")
	if err!=nil {
		log.Fatal(err)
	}
	peer.PrivateKey=privateKey

	publicKey, err := cmd.CommandCaptureOutputStdin(privateKey,"wg","pubkey")
	if err!=nil {
		log.Fatal(err)
	}
	peer.PublicKey=publicKey
	privateKeyFile, err := ioutil.TempFile("", "privatekey")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(privateKeyFile, strings.NewReader(privateKey)); err != nil {
		log.Fatal(err)
	}

	peer.PrivateKeyPath=privateKeyFile.Name();
	if err != nil {
		log.Fatal(err)
	}
	exitCode = cmd.Command("wg","set",wireguard.InterfaceName,"listen-port",strconv.Itoa(peer.ListenPort),"private-key",peer.PrivateKeyPath)
	cmd.ValidateCommand(exitCode)

	exitCode = cmd.Command("ip","link","set",wireguard.InterfaceName,"up")
	cmd.ValidateCommand(exitCode)

	fmt.Println("---> This node ID:",publicKey)
}

func (wireguard *WireGuardNetworkService) LinkPeer(peer MeshRemotePeer) {
	fmt.Println("---->New Peer discovered " + peer.PublicKey)
	allowedIPS:=strings.Join(peer.AllowedIPs,",")
	exitCode  := cmd.Command("wg","set",wireguard.InterfaceName,"peer",peer.PublicKey,"persistent-keepalive",strconv.Itoa(peer.KeepAlive),"allowed-ips",allowedIPS,"endpoint", peer.PublicIP+":"+strconv.Itoa(peer.PublicPort))
	cmd.ValidateCommand(exitCode)
	for _,subnet := range  peer.AllowedIPs {
		cmd.Command("ip","route","del",subnet,"dev",wireguard.InterfaceName)
		exitCode  = cmd.Command("ip","route","add",subnet,"dev",wireguard.InterfaceName)
		cmd.ValidateCommand(exitCode)
	}

}

func (wireguard *WireGuardNetworkService) UpdatePeer(peer MeshRemotePeer) {
	fmt.Println("---->Peer Update " + peer.PublicKey)
	wireguard.UnlinkPeer(peer)
	wireguard.LinkPeer(peer)
}

func (wireguard *WireGuardNetworkService) UnlinkPeer(peer MeshRemotePeer) {
	fmt.Println("---->Unlink Peer " + peer.PublicKey)
	exitCode  := cmd.Command("wg","set",wireguard.InterfaceName,"peer",peer.PublicKey,"remove")
	cmd.ValidateCommand(exitCode)

	for _,subnet := range  peer.AllowedIPs {
		cmd.Command("ip","route","del",subnet,"dev",wireguard.InterfaceName)
	}
}

func (wireguard *WireGuardNetworkService) DestroyNetworkDevice(peer MeshLocalPeer) {
	fmt.Println("---->Destroying wireguard device "+wireguard.InterfaceName)
	exitCode  := cmd.Command("ip","link","delete","dev",wireguard.InterfaceName)
	cmd.ValidateCommand(exitCode)

	exitCode  = cmd.Command("iptables","-D","FORWARD","-i",wireguard.InterfaceName ,"-j","ACCEPT")
	cmd.ValidateCommand(exitCode)

	defer os.Remove(peer.PrivateKeyPath)
}

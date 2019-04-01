package mesh

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"github.com/segator/wireguard-dynamic/cmd"
)
type WireGuardNetworkService struct {
	InterfaceMTU int
}


func NewWireGuardNetworkService() NetworkService {
	return &WireGuardNetworkService{
		InterfaceMTU: 1420,
	}
}

func (wireguard *WireGuardNetworkService) InitializeNetworkDevice(peer *MeshLocalPeer) {
	//Delete if exists
	cmd.Command("ip","link","delete","dev",peer.DeviceName)
	cmd.Command("iptables","-D","FORWARD","-i",peer.DeviceName ,"-j","ACCEPT")

	log.Println("---->Initializing (Wireguard) Device " + peer.DeviceName)


    exitCode  := cmd.Command("ip","link","add",peer.DeviceName,"type","wireguard")
    cmd.ValidateCommand(exitCode)

	exitCode  = cmd.Command("ip","address","add",peer.VPNIP+"/32","dev",peer.DeviceName)
	cmd.ValidateCommand(exitCode)

	exitCode = cmd.Command("ip","link","set","mtu",strconv.Itoa(wireguard.InterfaceMTU),"up","dev",peer.DeviceName)
	cmd.ValidateCommand(exitCode)

	exitCode  = cmd.Command("iptables","-A","FORWARD","-i",peer.DeviceName ,"-j","ACCEPT")
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
	exitCode = cmd.Command("wg","set",peer.DeviceName,"listen-port",strconv.Itoa(peer.ListenPort),"private-key",peer.PrivateKeyPath)
	cmd.ValidateCommand(exitCode)

	exitCode = cmd.Command("ip","link","set",peer.DeviceName,"up")
	cmd.ValidateCommand(exitCode)

	log.Println("---> This node ID:",publicKey)
}

func (wireguard *WireGuardNetworkService) LinkPeer(localPeer *MeshLocalPeer,peer *MeshRemotePeer) {
	log.Println("---->New Peer (Wireguard) discovered " + peer.PublicKey)
	allowedIPS:=strings.Join(peer.AllowedIPs,",")
	exitCode  := cmd.Command("wg","set",localPeer.DeviceName,"peer",peer.PublicKey,"persistent-keepalive",strconv.Itoa(peer.KeepAlive),"allowed-ips",allowedIPS,"endpoint", peer.PublicIP+":"+strconv.Itoa(peer.PublicPort))
	cmd.ValidateCommand(exitCode)
	for _,subnet := range  peer.AllowedIPs {
		cmd.Command("ip","route","del",subnet)
		exitCode  = cmd.Command("ip","route","add",subnet,"dev",localPeer.DeviceName)
		cmd.ValidateCommand(exitCode)
	}

}

func (wireguard *WireGuardNetworkService) UpdatePeer(localPeer *MeshLocalPeer,beforeUpdatePeer *MeshRemotePeer,afterUpdatePeer *MeshRemotePeer) {
	log.Println("---->Peer Update (Wireguard) " + beforeUpdatePeer.PublicKey)
	wireguard.UnlinkPeer(localPeer,beforeUpdatePeer)
	wireguard.LinkPeer(localPeer,afterUpdatePeer)
}

func (wireguard *WireGuardNetworkService) UnlinkPeer(localPeer *MeshLocalPeer,peer *MeshRemotePeer) {
	log.Println("---->Unlink Peer (Wireguard) " + peer.PublicKey)
	exitCode  := cmd.Command("wg","set",localPeer.DeviceName,"peer",peer.PublicKey,"remove")
	cmd.ValidateCommand(exitCode)

	for _,subnet := range  peer.AllowedIPs {
		cmd.Command("ip","route","del",subnet)
	}
}

func (wireguard *WireGuardNetworkService) DestroyNetworkDevice(peer *MeshLocalPeer) {
	log.Println("---->Destroying (Wireguard)  device "+peer.DeviceName)
	exitCode  := cmd.Command("ip","link","delete","dev",peer.DeviceName)
	cmd.ValidateCommand(exitCode)

	exitCode  = cmd.Command("iptables","-D","FORWARD","-i",peer.DeviceName ,"-j","ACCEPT")
	cmd.ValidateCommand(exitCode)

	defer os.Remove(peer.PrivateKeyPath)
}

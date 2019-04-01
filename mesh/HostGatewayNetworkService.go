package mesh

import (
	"encoding/json"
	"github.com/segator/wireguard-dynamic/cmd"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type HostGatewayNetworkService struct {
	backendNetworkService NetworkService
}


func NewHostGatewayNetworkService(backendNetworkService NetworkService) NetworkService {
	return &HostGatewayNetworkService{
		backendNetworkService: backendNetworkService,
	}
}

func (hostgw *HostGatewayNetworkService) InitializeNetworkDevice(peer *MeshLocalPeer) {
	//Delete if exists
	hostgw.backendNetworkService.InitializeNetworkDevice(peer)
}

func (hostgw *HostGatewayNetworkService) LinkPeer(localPeer *MeshLocalPeer,peer *MeshRemotePeer) {
	//Check every private IP if accesible to wg-dynamic rest api
	peer.HostGWMode=false
	peer.HostGWIp=""
	for _ , privateIP := range peer.PrivateIPs {
		restAPIURL := "http://" + privateIP + ":" + strconv.Itoa(peer.ApiListenPort) + "/status"
		timeout := time.Duration(1 * time.Second)
		client := http.Client{Timeout: timeout}
		resp, err:= client.Get(restAPIURL)
		if err==nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				bodyBytes, _ := ioutil.ReadAll(resp.Body)
				var restApiStatus APILocalPeerStatus
				if err :=json.Unmarshal(bodyBytes, &restApiStatus ); err != nil {
					log.Println("Invalid JSON decode",bodyBytes)
				}else{
					if restApiStatus.PublicKey == peer.PublicKey {
						exitCode := cmd.Command("ip","route","add",peer.AllowedIPs[0],"via",privateIP)
						if exitCode == 0 {
							cmd.Command("ip","route","del",peer.AllowedIPs[0])
							peer.HostGWMode=true
							peer.HostGWIp = privateIP
							break
						}

					}
				}
			}
		}
	}
	if peer.HostGWMode {
		log.Println("---->New Peer (Host-GW) discovered " + peer.PublicKey)
		for _,subnet := range  peer.AllowedIPs {
			cmd.Command("ip","route","del",subnet)
			exitCode := cmd.Command("ip","route","add",subnet,"via",peer.HostGWIp)
			cmd.ValidateCommand(exitCode)
		}
	}else{
		hostgw.backendNetworkService.LinkPeer(localPeer,peer)
	}
	}

func (hostgw *HostGatewayNetworkService) UpdatePeer(localPeer *MeshLocalPeer,beforeUpdatePeer *MeshRemotePeer,afterUpdatePeer *MeshRemotePeer) {
	if beforeUpdatePeer.HostGWMode {
		log.Println("---->Peer Update (Host-GW) " + beforeUpdatePeer.PublicKey)
		hostgw.UnlinkPeer(localPeer,beforeUpdatePeer)
		hostgw.LinkPeer(localPeer,afterUpdatePeer)
	}else{
		hostgw.backendNetworkService.UpdatePeer(localPeer,beforeUpdatePeer,afterUpdatePeer)
	}
}

func (hostgw *HostGatewayNetworkService) UnlinkPeer(localPeer *MeshLocalPeer,peer *MeshRemotePeer) {
	if peer.HostGWMode {
		log.Println("---->Unlink Peer (Host-GW) " + peer.PublicKey)
		for _,subnet := range  peer.AllowedIPs {
			cmd.Command("ip", "route", "del", subnet)
		}
	}else{
		hostgw.backendNetworkService.UnlinkPeer(localPeer,peer)
	}
}

func (hostgw *HostGatewayNetworkService) DestroyNetworkDevice(peer *MeshLocalPeer) {
	hostgw.backendNetworkService.DestroyNetworkDevice(peer)
}


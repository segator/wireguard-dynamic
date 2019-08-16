package mesh

import (
	"encoding/json"
	"github.com/segator/wireguard-dynamic/cmd"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
		if len(localPeer.PrivateIPs) > 0 && localPeer.PrivateIPs[0] != "none" {
		for _ , privateIP := range peer.PrivateIPs {
			if privateIP == "none" {
				break
			}
			if hostgw.checkStatus(privateIP,peer,1) {
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

	if peer.HostGWMode {
		allowedIPS:=strings.Join(localPeer.AllowedIPs,",")
		log.Println("New Peer (Host-GW) " + peer.PublicKey + " (privateIP=" + allowedIPS+" publicIP="+ peer.PublicIP+":"+strconv.Itoa(peer.PublicPort)+ ")")
		for _,subnet := range  peer.AllowedIPs {
			cmd.Command("ip","route","del",subnet)
			exitCode := cmd.Command("ip","route","add",subnet,"via",peer.HostGWIp)
			cmd.ValidateCommand(exitCode)
		}
	}else{
		hostgw.backendNetworkService.LinkPeer(localPeer,peer)
	}
	}

func (hostgw *HostGatewayNetworkService) checkStatus(ip string,peer *MeshRemotePeer,timeoutSeconds time.Duration) bool {
	result := false
	restAPIURL := "http://" + ip + ":" + strconv.Itoa(peer.ApiListenPort) + "/status"
	timeout := time.Duration(timeoutSeconds * time.Second)
	client := http.Client{Timeout: timeout}
	resp, err:= client.Get(restAPIURL)
	if err==nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			var restApiStatus APILocalPeerStatus
			if err := json.Unmarshal(bodyBytes, &restApiStatus); err != nil {
				log.Println("Invalid JSON decode", bodyBytes)
			} else {
				if restApiStatus.PublicKey == peer.PublicKey {
					result = true
				}
			}
		}
	}
	return result
}

func (hostgw *HostGatewayNetworkService) UpdatePeer(localPeer *MeshLocalPeer,beforeUpdatePeer *MeshRemotePeer,afterUpdatePeer *MeshRemotePeer) {
	if beforeUpdatePeer.HostGWMode {
		allowedIPS:=strings.Join(localPeer.AllowedIPs,",")
		log.Println("Peer Update (Host-GW) " + beforeUpdatePeer.PublicKey + " (privateIP=" + allowedIPS+" publicIP="+ beforeUpdatePeer.PublicIP+":"+strconv.Itoa(beforeUpdatePeer.PublicPort)+ ")")
		hostgw.UnlinkPeer(localPeer,beforeUpdatePeer)
		hostgw.LinkPeer(localPeer,afterUpdatePeer)
	}else{
		hostgw.backendNetworkService.UpdatePeer(localPeer,beforeUpdatePeer,afterUpdatePeer)
	}
}

func (hostgw *HostGatewayNetworkService) UnlinkPeer(localPeer *MeshLocalPeer,peer *MeshRemotePeer) {
	if peer.HostGWMode {
		allowedIPS:=strings.Join(localPeer.AllowedIPs,",")
		log.Println("Unlink Update (Host-GW) " + peer.PublicKey + " (privateIP=" + allowedIPS+" publicIP="+ peer.PublicIP+":"+strconv.Itoa(peer.PublicPort)+ ")")
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


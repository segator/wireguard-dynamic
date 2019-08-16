package mesh

import (
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
	"github.com/segator/wireguard-dynamic/retry"
)


type SimpleMeshService struct {
	repository PeerRepository
	network NetworkService
	ipPublic PublicIPRepository
	mesh Mesh
	localPeer *MeshLocalPeer
	peers *RemotePeersStored
	shutdown bool
}



func NewSimpleMeshService(repository PeerRepository,network NetworkService) MeshService {
	return &SimpleMeshService{
		repository: repository,
		network: network,
		ipPublic : NewPublicIPRepository(),
	}
}
func (meshService *SimpleMeshService) GetLocalPeer() MeshLocalPeer {
	return *meshService.localPeer
}
func (meshService *SimpleMeshService) Stop() {
	meshService.shutdown=true
	_,err :=retry.Do(func()  (interface{},*retry.RetryError){
		err := meshService.repository.Delete(meshService.mesh.MeshID,meshService.localPeer)
		if err!=nil {
			return nil,&retry.RetryError{true,err	}
		}
		return nil,nil
	})
	if err!=nil {
		log.Panic(err)
	}

	for _,peer := range meshService.peers.remotePeers{
		meshService.network.UnlinkPeer(meshService.localPeer,peer)
	}
	meshService.network.DestroyNetworkDevice(meshService.localPeer)
}

func ( meshService *SimpleMeshService) CreateMesh() Mesh{
	bucketID,err :=retry.Do(func()  (interface{},*retry.RetryError){
		bucketID,err := meshService.repository.CreateBucket()
		if err!=nil {
			return nil,&retry.RetryError{true,err	}
		}
		return bucketID,nil
	})
	if err!=nil {
		log.Panic(err)
	}

	return Mesh{
		MeshID: bucketID.(string),
	}
}
func  (meshService *SimpleMeshService) JoinMesh(mesh Mesh,localPeer MeshLocalPeer){
	meshService.mesh=mesh
	meshService.localPeer= &localPeer
	//load peers
	peersStoredInterface,err :=retry.Do(func()  (interface{},*retry.RetryError){
		peersStored,err := meshService.repository.FindAll(mesh.MeshID)
		if err!=nil {
			return nil,&retry.RetryError{true,err	}
		}
		return peersStored,nil
	})
	if err!=nil {
		log.Fatal(err)
	}
	peersStored := peersStoredInterface.([]*MeshRemotePeer)
	meshService.peers =&RemotePeersStored{
		remotePeers:[]*MeshRemotePeer{},
	}
	if len(peersStored)==0{
		log.Println("--> This is the first node on the mesh!!")
	}
	if meshService.localPeer.AutoVPNIP {
		meshService.localPeer.VPNIP = findFreeIP(peersStored)
	}
	meshService.localPeer.AllowedIPs = append(meshService.localPeer.AllowedIPs,meshService.localPeer.VPNIP+"/32")

	//initialize device
	meshService.network.InitializeNetworkDevice(&localPeer)

	go monitorPeers(meshService)
	go monitorPublicIP(meshService)


}

func calculatePublicIP(meshService  *SimpleMeshService) string {
	publicIP,err :=retry.Do(func()  (interface{},*retry.RetryError){
		publicIP, err := meshService.ipPublic.getPublicIP()
		if err!=nil {
			return nil,&retry.RetryError{true,err	}
		}
		return publicIP,nil
	})
	if err!=nil {
		log.Fatal(err)
	}
	return publicIP.(string)
}
func monitorPeers(meshService *SimpleMeshService){
   for {
   	   if meshService.shutdown {
   	   	return
	   }
	   peersInterface,err :=retry.Do(func()  (interface{},*retry.RetryError){
		   peersStored,err := meshService.repository.FindAll(meshService.mesh.MeshID)
		   if err!=nil {
		   	   log.Println("error getting nodes",err)
			   return nil,&retry.RetryError{true,err	}
		   }
		   return peersStored,nil
	   })
	   if err!=nil {
		   log.Fatal(err)
	   }
	   peers := peersInterface.([]*MeshRemotePeer)
	   if err==nil {
		   //Find Deletes
		   toUnBind:= []*MeshRemotePeer{}
		   newPeerList:= []*MeshRemotePeer{}
		   for _ , remotePeerBefore := range meshService.peers.remotePeers {
			   found :=false
			   for _, remotePeer := range peers {
				   if remotePeer.Compare(*remotePeerBefore) {
					   found  = true
				   }
			   }
			   if !found {
				   toUnBind = append(toUnBind,remotePeerBefore)
			   }else{
				   newPeerList = append(newPeerList,remotePeerBefore)
			   }
		   }
		   for _, unlinkPeer := range toUnBind {
			   meshService.network.UnlinkPeer(meshService.localPeer,unlinkPeer)
		   }
		   meshService.peers.remotePeers = newPeerList


	   	   //Find updates and Inserts
		   for _, remotePeer := range peers {
		   		found := false
		   	    for storeIndex , remotePeerBefore := range meshService.peers.remotePeers {
		   	    	if remotePeer.Compare(*remotePeerBefore) {
						found=true
						if remotePeer.Version > remotePeerBefore.Version {
							meshService.peers.remotePeers[storeIndex] = remotePeer
							meshService.network.UpdatePeer(meshService.localPeer,remotePeerBefore,remotePeer)
						}
					}
				}
		   	    //Add new Peers
		   	    if !found {
					if meshService.localPeer.PublicKey != remotePeer.PublicKey {
						meshService.network.LinkPeer(meshService.localPeer,remotePeer)
						meshService.peers.remotePeers = append(meshService.peers.remotePeers,remotePeer)
					}
				}
		   }


	   }
	   time.Sleep(time.Duration(meshService.localPeer.KeepAlive) * time.Second)
   }
}
func monitorPublicIP(meshService *SimpleMeshService){
	//i:=1
	for {
		if meshService.shutdown {
			return
		}
		if meshService.localPeer.AutoPublicIP {
			newPublicIP := calculatePublicIP(meshService)
			/*if(i%2==0){
				newPublicIP="24.1.4.1"
			}
			i=i+1;*/
			if(meshService.localPeer.PublicIP != newPublicIP){
				log.Println("Detect new Public IP("+ newPublicIP + ") Updating nodes configuration...")
				meshService.localPeer.PublicIP =newPublicIP
				meshService.localPeer.Version++
			}
		}
		_,err :=retry.Do(func()  (interface{},*retry.RetryError){
			err := meshService.repository.Store(meshService.mesh.MeshID,meshService.localPeer.MeshRemotePeer)
			if err!=nil {
				log.Println("Error executing store",err)
				return nil,&retry.RetryError{true,err	}
			}
			return nil,nil
		})
		if err!=nil {
			log.Panic(err)
		}



		//every time we modify this node update the version
		//meshService.localPeer.version++
		time.Sleep(time.Duration(meshService.localPeer.KeepAlive) * time.Second)
	}
}

func findFreeIP(peers []*MeshRemotePeer) string{
	usedIPS := []string{}
	for _, remotePeer := range peers {
		ip,_,err := net.ParseCIDR(remotePeer.VPNIP+"/32")
		if err ==nil {
			usedIPS = append(usedIPS,ip.String())
		}
	}
	newIP := "12.1.1.1"
	for StringSliceContains(usedIPS,newIP) {
		numRand := rand.Intn(254 - 1) + 1
		newIP = "12.1.1." + strconv.Itoa(numRand)
	}
	return newIP
}

func StringSliceContains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}


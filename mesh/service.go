package mesh

import (
	"net/http"
)

type MeshService interface {
	CreateMesh() Mesh
	JoinMesh(mesh Mesh,meshPeer MeshLocalPeer)
	Stop()
	GetLocalPeer() MeshLocalPeer

}

type NetworkService interface {
	InitializeNetworkDevice(peer *MeshLocalPeer)
	LinkPeer(localPeer *MeshLocalPeer,peer *MeshRemotePeer)
	UpdatePeer(localPeer *MeshLocalPeer,beforeUpdatePeer *MeshRemotePeer,afterUpdatePeer *MeshRemotePeer)
	UnlinkPeer(localPeer *MeshLocalPeer,peer *MeshRemotePeer)
	DestroyNetworkDevice(peer *MeshLocalPeer)
}

type RestAPIService interface {
    Listen()
	GetStatus(writer http.ResponseWriter, request *http.Request)
}

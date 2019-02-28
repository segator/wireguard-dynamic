package mesh

type MeshService interface {
	CreateMesh() Mesh
	JoinMesh(mesh Mesh,meshPeer MeshLocalPeer)
	Stop()
}

type NetworkService interface {
	InitializeNetworkDevice(peer *MeshLocalPeer)
	LinkPeer(localPeer *MeshLocalPeer,peer MeshRemotePeer)
	UpdatePeer(localPeer *MeshLocalPeer,peer MeshRemotePeer)
	UnlinkPeer(localPeer *MeshLocalPeer,peer MeshRemotePeer)
	DestroyNetworkDevice(peer MeshLocalPeer)
}


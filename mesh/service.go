package mesh

type MeshService interface {
	CreateMesh() Mesh
	JoinMesh(mesh Mesh,meshPeer MeshLocalPeer)
	Stop()
}

type NetworkService interface {
	InitializeNetworkDevice(peer *MeshLocalPeer)
	LinkPeer(peer MeshRemotePeer)
	UpdatePeer(peer MeshRemotePeer)
	UnlinkPeer(peer MeshRemotePeer)
	DestroyNetworkDevice(peer MeshLocalPeer)
}


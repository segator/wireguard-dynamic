package mesh


type PeerRepository interface{
	CreateBucket() (string,error)
	FindAll(bucket string) ([]*MeshRemotePeer,error)
	Store(bucket string,peer MeshRemotePeer) error
	Update(bucket string,peer MeshRemotePeer) error
	Delete(bucket string, peer *MeshLocalPeer) error
}

type PublicIPRepository interface{
	getPublicIP() (string,error)
}

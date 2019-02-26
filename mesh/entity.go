package mesh

type Mesh struct {
	MeshID string `json:"mesh_id"`
}

type MeshRemotePeer struct {
	version   int `json:"version"`
	PublicIP  string `json:"public_ip"`
	AllowedIPs []string `json:"allowed_ips"`
	PublicPort int `json:"public_port"`
	VPNIP string `json:"vpn_ip"`
	PublicKey string  `json:"public_key"`
	KeepAlive int `json:"keepalive"`
}

type MeshLocalPeer struct {
	MeshRemotePeer
	AutoPublicIP bool
	AutoVPNIP bool
	ListenPort int `json:"listen_port"`
	PrivateKey string  `json:"private_key"`
	PrivateKeyPath string  `json:"private_key_path"`
}

type RemotePeersStored struct {
	remotePeers []*MeshRemotePeer
}

func (a MeshRemotePeer) Compare(b MeshRemotePeer) bool{
	return a.PublicKey == b.PublicKey
}





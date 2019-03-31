package mesh

type Mesh struct {
	MeshID string `json:"mesh_id"`
}

type MeshRemotePeer struct {
	Version   int `json:"version"`
	PublicIP  string `json:"public_ip"`
	AllowedIPs []string `json:"allowed_ips"`
	PublicPort int `json:"public_port"`
	VPNIP string `json:"vpn_ip"`
	PublicKey string  `json:"public_key"`
	KeepAlive int `json:"keepalive"`
	ListenPort int `json:"listen_port"`
	PrivateIPs []string `json:"private_ips"`
	ApiListenPort int `json:"api_listen_port"`

	HostGWMode bool
	HostGWIp string

}

type MeshLocalPeer struct {
	MeshRemotePeer
	AutoPublicIP bool
	AutoVPNIP bool
	DeviceName string
	PrivateKey string  `json:"private_key"`
	PrivateKeyPath string  `json:"private_key_path"`
}

type APILocalPeerStatus struct {
	PublicKey string  `json:"public_key"`
}


type RemotePeersStored struct {
	remotePeers []*MeshRemotePeer
}

func (a MeshRemotePeer) Compare(b MeshRemotePeer) bool{
	return a.PublicKey == b.PublicKey
}






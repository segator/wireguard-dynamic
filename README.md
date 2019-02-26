# wireguard-dynamic
Wireguard Auto Discovery Peers, allow you to deploy Wireguard networks with a single command.
Add/update/remove peers.
Check peer status and conectivity.
TO achieve this we use key-value stores, for now we are using kvdb.io (free service)

## Install
To install first you need to  install wireguard, check the official documentation.
https://www.wireguard.com/install/

## Build
to Build you have 2 ways, build yourself go application or build docker

* GO

Make sure you have golang installed >1.11
```bash
go get  https://github.com/segator/wireguard-dynamic
cd $GOPATH/bin
./wireguard-dynamic --help
```

* Docker

```bash
git clone https://github.com/segator/wireguard-dynamic
cd wireguard-dynamic
docker build -t segator/wireguard-dynamic . 
```

## Run
To run you can use the binary generated by go build or docker image

Example with docker
```bash
#First lets create a new network
NETWORK_TOKEN=$(docker run -it --rm --net=host --privileged  segator/wireguard-dynamic init)

#On Node1/Node 2/Node 3.. whatever node you want
docker run -d --restart=always \
           --name wireguard \
           --net=host \
           --privileged \
           segator/wireguard-dynamic join --token $NETWORK_TOKEN --listen-port 31111
           
 #wait 5-10 seconds, then the nodes will be auto detected and interconected.
 #to get the ip of the node wireguard devices
 ifconfig wg0
```

## Features
* [X] Peer auto discovery
* [X] Dynamic IP support(Ready for home-user connections)
* [ ] NAT passthorugh(UPNP and UDP Hole punching using STUN servers)
* [ ] Support other stores,typical K/V(etcd,consul,zookeeper...)
* [ ] Support other public kv: https://keyvalue.xyz/
* [ ] Kubernetes CNI


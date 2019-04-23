package mesh

import (
	"encoding/base64"
	"encoding/json"
	"github.com/abronan/valkeyrie"
	"github.com/abronan/valkeyrie/store"
	"github.com/abronan/valkeyrie/store/boltdb"
	"github.com/abronan/valkeyrie/store/consul"
	"github.com/abronan/valkeyrie/store/dynamodb"
	"github.com/abronan/valkeyrie/store/etcd/v3"
	"github.com/abronan/valkeyrie/store/redis"
	"github.com/abronan/valkeyrie/store/zookeeper"
	"log"
	"strconv"
	"time"
)

type ValkeyrieKVType = store.Backend
type ValkeyrieConfig = store.Config

const (
	CONSUL ValkeyrieKVType = "consul"
	ETCDV3 ValkeyrieKVType = "etcdv3"
	ZK ValkeyrieKVType = "zk"
	BOLTDB ValkeyrieKVType = "boltdb"
	REDIS ValkeyrieKVType = "redis"
	DYNAMODB ValkeyrieKVType = "dynamodb"
)

type ValkeyrieRepository struct {
	store store.Store
}



func NewKValkeyrieRepository(kvType ValkeyrieKVType,addr []string,config *ValkeyrieConfig) PeerRepository {
	switch kvType {
	case CONSUL:
		consul.Register()
	case ETCDV3:
		etcdv3.Register()
	case ZK:
		zookeeper.Register()
	case REDIS:
		redis.Register()
	case BOLTDB:
		boltdb.Register()
	case DYNAMODB:
		dynamodb.Register()
	}
	kv, err := valkeyrie.NewStore(kvType,addr,config)
	if err!= nil {
		log.Fatal("Cannot create store "+kvType)
	}
	return &ValkeyrieRepository{
		store:kv,
	}
}

func (repository *ValkeyrieRepository) CreateBucket() (string,error) {
	bucketName := strconv.FormatInt(time.Now().UnixNano(),10)
	err := repository.store.Put(bucketName,nil,&store.WriteOptions{IsDir:true})
	return bucketName,err
}


func (repository *ValkeyrieRepository) FindAll(bucket string) ([]*MeshRemotePeer,error) {
	entries, err := repository.store.List(bucket,&store.ReadOptions{})
	var peersList []*MeshRemotePeer
	if err != nil {
		return  peersList,err
	}
	for _,pair := range entries {
			var peer MeshRemotePeer
			if err=json.Unmarshal([]byte(pair.Value),&peer); err!= nil {
				return nil,&errorStringValkyrie{
					s: "invalid JSON decode",
				}
			}
			peersList =append(peersList, &peer)
		}
		return peersList, nil
	}


func  (repository *ValkeyrieRepository) Store(bucket string,peer MeshRemotePeer) error {
	bytesJSON, err := json.Marshal(peer)
	if(err!=nil){
		return err
	}
	publicKeyB64 := base64.StdEncoding.EncodeToString([]byte(peer.PublicKey))
	err = repository.store.Put(bucket+"/"+publicKeyB64,bytesJSON,&store.WriteOptions{TTL:time.Duration(peer.KeepAlive *2)*time.Second})
	if err!=nil {
		return err
	}
	return nil
}

func (repository *ValkeyrieRepository) Delete(bucket string, peer *MeshLocalPeer) error {
	publicKeyB64 := base64.StdEncoding.EncodeToString([]byte(peer.PublicKey))
	err := repository.store.Delete(bucket+"/"+publicKeyB64)
	return err
}
func  (repository *ValkeyrieRepository) Update(bucket string,peer MeshRemotePeer) error{
	return repository.Store(bucket,peer)
}
type errorStringValkyrie struct {
	s string
}

func (e *errorStringValkyrie) Error() string {
	return e.s
}
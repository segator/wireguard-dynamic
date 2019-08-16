package mesh

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type KVDBRepository struct {
	ServerURL url.URL
}



func NewKVDBRepository() PeerRepository {
	kvdbURL,err := url.Parse("https://kvdb.io")
	if err!=nil {
		log.Fatal(err)
	}
	return &KVDBRepository{
		ServerURL:*kvdbURL,
	}
}

func (repository *KVDBRepository) CreateBucket() (string,error) {
	response := ""
	client := http.Client{Timeout: time.Duration(5 * time.Second)}
	resp, err := client.Post(repository.ServerURL.String(),"application/json",nil)
	if err!=nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		response = string(bodyBytes)
		return response,nil
	}else{
		return  "",&errorStringKVDB{
			s: "invalid status code",
		}
	}


}


func (repository *KVDBRepository) FindAll(bucket string) ([]*MeshRemotePeer,error) {
	client := http.Client{Timeout: time.Duration(5 * time.Second)}
	resp, err:= client.Get(repository.ServerURL.String()+"/"+bucket+"/?format=json&values=true")
	if err!=nil {
		return nil,err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		var elements [][]string
		if err :=json.Unmarshal(bodyBytes, &elements ); err != nil {
			return nil,&errorStringKVDB{
				s: "invalid JSON decode",
			}
		}

		var peersList []*MeshRemotePeer
		for _ , element := range elements {
			var peer MeshRemotePeer
			if err=json.Unmarshal([]byte(element[1]),&peer); err!= nil {
				return nil,&errorStringKVDB{
					s: "invalid JSON decode",
				}
			}
			peersList =append(peersList, &peer)
		}
		return peersList, nil
	}else{
		return  nil,&errorStringKVDB{
			s: "invalid status code",
		}
	}
}

func  (repository *KVDBRepository) Store(bucket string,peer MeshRemotePeer) error {
	bytesJSON, err := json.Marshal(peer)
	if(err!=nil){
		return err
	}
	publicKeyB64 := base64.StdEncoding.EncodeToString([]byte(peer.PublicKey))
	url := repository.ServerURL.String()+"/"+bucket+"/"+publicKeyB64+"?ttl=" + strconv.Itoa(peer.KeepAlive *2)
	client := http.Client{Timeout: time.Duration(5 * time.Second)}
	resp, err := client.Post(url,"application/json",bytes.NewReader(bytesJSON))
	if err!=nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Print(url)
		return  &errorStringKVDB{
			s: "invalid status code: "+strconv.Itoa(resp.StatusCode),
		}
	}


	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (repository *KVDBRepository) Delete(bucket string, peer *MeshLocalPeer) error {
	publicKeyB64 := base64.StdEncoding.EncodeToString([]byte(peer.PublicKey))
	client := &http.Client{}
	url := repository.ServerURL.String()+"/"+bucket+"/"+publicKeyB64
	request, err:= http.NewRequest(http.MethodDelete,url,nil)
	if err!=nil {
		log.Fatal(err)
	}
	resp, err :=client.Do(request)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusAccepted {
		return nil
	}else{
		return  &errorStringKVDB{
			s: "invalid status code "+resp.Status,
		}
	}
}
func  (repository *KVDBRepository) Update(bucket string,peer MeshRemotePeer) error{
	return repository.Store(bucket,peer)
}
type errorStringKVDB struct {
	s string
}

func (e *errorStringKVDB) Error() string {
	return e.s
}
package mesh

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type RestAPI struct {
	MeshService MeshService
	ListenPort int

}

func NewRestAPIService(listenPort int, meshService MeshService) RestAPIService {
	return &RestAPI{
		ListenPort: listenPort,
		MeshService:meshService,
	}
}

func (api *RestAPI) Listen() {
	router := mux.NewRouter()
	router.HandleFunc("/status", api.GetStatus).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(api.ListenPort), router))
}

func (api *RestAPI) GetStatus(writer http.ResponseWriter, request *http.Request){
	stats := &APILocalPeerStatus{
		PublicKey: api.MeshService.GetLocalPeer().PublicKey,
	}
	json.NewEncoder(writer).Encode(stats)
}

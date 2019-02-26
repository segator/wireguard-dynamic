package mesh

import (
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type publicIPRepoProviders struct{
	providers []string
}



func NewPublicIPRepository() PublicIPRepository {
	var providers = []string{"https://api.ipify.org/","http://ifconfig.io/ip","http://ipecho.net/plain","http://icanhazip.com","https://ifconfig.me/ip"}
	return &publicIPRepoProviders{
		providers: providers,

	}
}

func (repoProvider publicIPRepoProviders) getPublicIP() (string, error) {
	numRand := rand.Intn(len(repoProvider.providers) - 0)

	resp, err:= http.Get(repoProvider.providers[numRand])
	if err!=nil {
		return "",err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		publicIP := strings.TrimSpace(string(bodyBytes))
		ipParsed :=  net.ParseIP(publicIP)
		if ipParsed ==nil {
			return "",&errorString{
				s: "Invalid IP " + publicIP,
			}
		}

		return publicIP,nil
	}else{
		return "",&errorString{
			s: "Invalid response "+  strconv.Itoa(resp.StatusCode),
		}
	}
}


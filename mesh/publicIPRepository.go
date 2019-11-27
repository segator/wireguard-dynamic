package mesh

import (
	"context"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type publicIPRepoProviders struct{
	providers []string
	transport *http.Transport
}



func NewPublicIPRepository() PublicIPRepository {
	var providers = []string{"https://api.ipify.org/","http://ifconfig.io/ip","http://ipecho.net/plain","http://icanhazip.com","https://ifconfig.me/ip","https://myexternalip.com/raw"}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
	}

	var MyTransport = &http.Transport{
		DialContext: dialer.DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	MyTransport.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp4", addr)
	}
	return &publicIPRepoProviders{
		providers: providers,
		transport: MyTransport,
	}
}

func (repoProvider publicIPRepoProviders) getPublicIP() (string, error) {
	numRand := rand.Intn(len(repoProvider.providers) - 0)
	client := http.Client{Timeout: time.Duration(5 * time.Second)}
	resp, err:= client.Get(repoProvider.providers[numRand])
	if err!=nil {
		return "",err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		publicIP := strings.TrimSpace(string(bodyBytes))
		ipParsed :=  net.ParseIP(publicIP)
		if ipParsed ==nil {
			return "",&errorStringPublicIP{
				s: "Invalid IP " + publicIP,
			}
		}

		return publicIP,nil
	}else{
		return "",&errorStringPublicIP{
			s: "Invalid response "+  strconv.Itoa(resp.StatusCode),
		}
	}
}

type errorStringPublicIP struct {
	s string
}

func (e *errorStringPublicIP) Error() string {
	return e.s
}

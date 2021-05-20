package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/getfider/fider/app/pkg/env"
)

//RunPing checks if Fider Server is running and is healthy
//Returns an exitcode, 0 for OK and 1 for ERROR
func RunPing() int {
	protocol := "http://"
	if env.Config.TLS.Certificate != "" || env.Config.TLS.Automatic {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		protocol = "https://"
	}

	resp, err := http.Get(protocol + "localhost:3000/_health")
	if err != nil {
		fmt.Printf("Request failed with: %s\n", err)
		return 1
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
		return 1
	}

	return 0
}

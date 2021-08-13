package test

import (
	//"bufio"
	"encoding/csv"
	"io"
	"io/ioutil"
	"net"
	"os"

	//"encoding/csv"
	"fmt"
	"github.com/stretchr/testify/assert"
	//"io/ioutil"
	"log"
	//"net"
	"net/http"
	"testing"
	//_ "github.com/lib/pq"
	//_ "github.com/go-sql-driver/mysql"
	//"github.com/t94j0/nmap"
	//"github.com/reiver/go-telnet"
)

func TestInternalInbound(t *testing.T) {
	csvfile, err := os.Open("az-lz-test_data/VMs.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	//Parse the file
	r := csv.NewReader(csvfile)
	endpoints, err := r.Read()
	if err != nil && err != io.EOF{
		log.Fatal(err)
	}
	for _, v := range endpoints {
		req, err := http.NewRequest("GET", v, nil)
		if err != nil {
			log.Fatal("Error reading request. ", err)
		}
		clientR := &http.Client{}
		resp, err := clientR.Do(req)
		if err != nil {
			log.Fatal("Error doing request. ", err)
		}
		assert.Equal(t, 200 , resp.StatusCode, "Failed to connect")
	}
}
func TestInternalOutbound(t *testing.T) {
	//3rd element is what i think is confluent IP. connection refused. confluent nic: 172.20.140.4
	//endpoints := [4]string{"http://gitlab.com/", "http://hub.docker.com", "http://registry.terraform.io", "http://docker.io"}
	acceptedStatuses := make(map[int]int)
	acceptedStatuses[200] = 200
	acceptedStatuses[403] = 403
	acceptedStatuses[470] = 470
	csvfile, err := os.Open("az-lz-test_data/InternalOutboundHTTPWhole.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	//Parse the file
	r := csv.NewReader(csvfile)
	record, err := r.Read()
	if err != nil && err != io.EOF{
		log.Fatal(err)
	}
	for _, val := range record{
		//record, err := r.Read()

		req, err := http.NewRequest("GET", "http://" + val, nil)
		if err != nil {
			log.Fatal("Error reading request. ", err)
		}
		fmt.Printf("Testing endpoint: %v\n", val)
		clientR := &http.Client{}
		resp, err := clientR.Do(req)
		if err != nil {
			log.Fatalf("Error doing request. %v\n ", err)
		}
		assert.Less(t,  resp.StatusCode, 500, resp.Status)

	}
}
func TestNetworkingAKSKeyVault (t *testing.T) {
	csvfile, err := os.Open("az-lz-test_data/aksKeyVault.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	//Parse the file
	r := csv.NewReader(csvfile)
	endpoints, err := r.Read()
	if err != nil && err != io.EOF{
		log.Fatal(err)
	}
	for _, val := range endpoints{
		req, err := http.NewRequest("GET", val, nil)
		if err != nil {
			log.Fatal("Error reading request. ", err)
		}
		fmt.Printf("Testing endpoint: %v\n", val)
		clientR := &http.Client{}
		resp, err := clientR.Do(req)
		if err != nil {
			log.Fatalf("Error doing request. %v\n ", err)
		}
		assert.Less(t, resp.StatusCode, 500, resp.Status)

	}
}
func TestNetworkingAKSControlPlane(t *testing.T){
	// web.yto22prdstr03a.store.core.windows.net
	// first is storage acc
	//endpoints := [5]string{"http://20.150.31.228:80","http://52.139.9.246:80", "http://aks-dev-cc-tracktrace-001-e4537629.hcp.canadacentral.azmk8s.io", "http://aks-qa-cc-tracktrace-001-e7169da1.hcp.canadacentral.azmk8s.io", "https://kv-dev-cc-main-001.vault.azure.net/"}
	csvfile, err := os.Open("az-lz-test_data/aksControlPlane.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	//Parse the file
	r := csv.NewReader(csvfile)
	endpoints, err := r.Read()
	if err != nil && err != io.EOF{
		log.Fatal(err)
	}
	for _, val := range endpoints{
		req, err := http.NewRequest("GET", val, nil)
		if err != nil {
			log.Fatal("Error reading request. ", err)
		}
		fmt.Printf("Testing endpoint: %v\n", val)
		clientR := &http.Client{}
		resp, err := clientR.Do(req)
		if err != nil {
			log.Fatalf("Error doing request. %v\n ", err)
		}
		assert.Less(t, resp.StatusCode, 500, resp.Status)

	}
}

func TestStorageAccounts(t *testing.T) {
	ips, err := net.LookupIP("ff78dee1a871640d5a7787b.blob.core.windows.net")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		os.Exit(1)
	}
	for _, ip := range ips {
		stringIp := ip.String()
		fmt.Printf("Sotrage Account. IN A %s\n", ip.String())
		tcpAddr, err := net.ResolveTCPAddr("tcp", stringIp + ":80")
		if err != nil {
			fmt.Printf("Error1: %v\n", err)
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		defer conn.Close()
		if err != nil {
			fmt.Printf("Error2: %v\n", err)
		}

		_, err = conn.Write([]byte("HEAD / TCP/1.0\r\n\r\n"))
		if err != nil {
			fmt.Printf("Error3: %v\n", err)
		}
		//result, err := readFully(conn)
		result, err := ioutil.ReadAll(conn)
		if err != nil {
			fmt.Printf("Error4: %v\n", err)
		}
		assert.Contains(t, string(result), "400 Bad Request")
		fmt.Println(string(result))
	}
}
func TestNetworkingKafkaEndpoints(t *testing.T) {
	//kafka 172.20.140.4:9092
	_, err := net.Dial("tcp", "172.20.140.4:9092")
	assert.Equal(t, nil, err, err)
}



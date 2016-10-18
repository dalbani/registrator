package etcd

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/coreos/go-etcd/etcd"
	"github.com/gliderlabs/registrator/bridge"
	"crypto/x509"
	"crypto/tls"
)

func init() {
	bridge.Register(new(Factory), "etcd2")
}

type Factory struct{}

func (f *Factory) New(uri *url.URL) bridge.RegistryAdapter {
	cert := os.Getenv("ETCD_CERT_FILE")
	key := os.Getenv("ETCD_KEY_FILE")
	cacert := os.Getenv("ETCD_CA_CERT_FILE")

	urls := make([]string, 0)

	var client *etcd.Client

	// Assuming https
	if cacert != "" {
		urls = append(urls, "https://"+uri.Host)

		// Assuming Client authentication if tlskey and tlspem is set
		if key != "" && cert != "" {
			var err error
			if client, err = etcd.NewTLSClient(urls, cert, key, cacert); err != nil {
				log.Fatalf("etcd: failure to connect: %s", err)
			}
		} else {
			client = etcd.NewClient(urls)
			ca, err := ioutil.ReadFile(cacert)
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(ca)
			tr := &http.Transport{
				TLSClientConfig:    &tls.Config{RootCAs: caCertPool},
				DisableCompression: true,
			}
			client.SetTransport(tr)
		}
	} else {
		urls = append(urls, "http://"+uri.Host)
		client = etcd.NewClient(urls)
	}

	return &EtcdAdapter{client: client, path: uri.Path}
}

type EtcdAdapter struct {
	client *etcd.Client

	path string
}

func (r *EtcdAdapter) Ping() error {
	r.syncEtcdCluster()

	var err error
	rr := etcd.NewRawRequest("GET", "version", nil, nil)
	_, err = r.client.SendRequest(rr)

	if err != nil {
		return err
	}
	return nil
}

func (r *EtcdAdapter) syncEtcdCluster() {
	var result bool
	result = r.client.SyncCluster()

	if !result {
		log.Println("etcd: sync cluster was unsuccessful")
	}
}

func (r *EtcdAdapter) Register(service *bridge.Service) error {
	r.syncEtcdCluster()

	path := r.path + "/" + service.Name + "/" + service.ID
	port := strconv.Itoa(service.Port)
	addr := net.JoinHostPort(service.IP, port)

	var err error
	_, err = r.client.Set(path, addr, uint64(service.TTL))

	if err != nil {
		log.Println("etcd: failed to register service:", err)
	}
	return err
}

func (r *EtcdAdapter) Deregister(service *bridge.Service) error {
	r.syncEtcdCluster()

	path := r.path + "/" + service.Name + "/" + service.ID

	var err error
	_, err = r.client.Delete(path, false)

	if err != nil {
		log.Println("etcd: failed to deregister service:", err)
	}
	return err
}

func (r *EtcdAdapter) Refresh(service *bridge.Service) error {
	return r.Register(service)
}

func (r *EtcdAdapter) Services() ([]*bridge.Service, error) {
	return []*bridge.Service{}, nil
}

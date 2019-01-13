package backend

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/vault/api"
)

const (
	vaultWireyPrefix = "wirey"
)

type VaultBackend struct {
	client *api.Client
}

func NewVaultBackend(endpoint string) (*VaultBackend, error) {
	conf := api.DefaultConfig()
	if len(endpoint) > 0 {
		conf.Address = endpoint
	}
	cli, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	mounts, err := cli.Sys().ListMounts()
	if err != nil {
		return nil, err
	}
	v, ok := mounts[fmt.Sprintf("%s/", vaultWireyPrefix)]
	if ok {
		log.Println(v)
		return &VaultBackend{
			client: cli,
		}, nil
	}
	fmt.Println("wirey/ doesn't exist yet as a backend")

	options := map[string]string{
		"version": "1",
	}
	wireyKV := &api.MountInput{
		Type:        "kv",
		Description: "KV esclusive for Wirey",
		SealWrap:    false,
		Local:       false,
		Options:     options,
		Config: api.MountConfigInput{
			DefaultLeaseTTL: "0",
			MaxLeaseTTL:     "0",
			ForceNoCache:    false,
		},
	}
	err = cli.Sys().Mount(vaultWireyPrefix, wireyKV)
	if err != nil {
		return nil, err
	}
	return &VaultBackend{
		client: cli,
	}, nil
}

func (v *VaultBackend) Join(ifname string, p Peer) error {
	peer := peerToMap(p)
	log.Println(peer)

	written, err := v.client.Logical().Write(fmt.Sprintf("%s/%s/%s", vaultWireyPrefix, ifname, base64.StdEncoding.EncodeToString(p.PublicKey)), peer)
	log.Println(written)
	if err != nil {
		return err
	}
	return nil
}

func (v *VaultBackend) GetPeers(ifname string) ([]Peer, error) {
	data, err := v.client.Logical().List(fmt.Sprintf("%s/%s", vaultWireyPrefix, ifname))
	if err != nil {
		return nil, err
	}
	log.Println(data)
	pKeys := []string{}
	if data == nil || data.Data == nil {
		return nil, nil
	}
	for _, v := range data.Data["keys"].([]interface{}) {
		pKeys = append(pKeys, v.(string))
	}
	log.Println(pKeys)
	peers := make([]Peer, 0, 0)
	for _, p := range pKeys {
		raw, err := v.client.Logical().Read(fmt.Sprintf("%s/%s/%s", vaultWireyPrefix, ifname, p))
		if err != nil {
			return nil, err
		}
		if raw == nil || raw.Data == nil {
			return nil, nil
		}
		peer, err := mapToPeer(raw.Data)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}
	return peers, nil
}

func peerToMap(p Peer) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	m["publicKey"] = base64.StdEncoding.EncodeToString(p.PublicKey)
	m["IP"] = p.IP.String()
	m["endpoint"] = p.Endpoint
	return m
}

func mapToPeer(raw map[string]interface{}) (Peer, error) {
	ip := net.ParseIP(raw["IP"].(string))
	pKey, err := base64.StdEncoding.DecodeString(raw["publicKey"].(string))
	if err != nil {
		return Peer{}, err
	}
	p := Peer{
		PublicKey: pKey,
		IP:        &ip,
		Endpoint:  raw["endpoint"].(string),
	}
	return p, nil
}

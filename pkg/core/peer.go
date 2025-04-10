package core

type Peer struct {
	Name       string
	PrivateKey string
	Address    string

	PublicKey    string
	AllowedIPs   string
	Endpoint     string
	PresharedKey string

	PersistentKeepalive int
	RouteAllowedIPs     bool
}

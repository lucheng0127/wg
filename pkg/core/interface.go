package core

type Iface struct {
	PublicKey string

	PrivateKey string
	Address    string
	ListenPort int
	DNS        string
	MTU        int

	FwMark     string
	Table      string
	PreUp      string
	PostUp     string
	PreDown    string
	PostDown   string
	SaveConfig bool
}

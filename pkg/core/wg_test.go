package core

import (
	"fmt"
	"testing"
)

func TestWG_UP(t *testing.T) {
	type fields struct {
		Interface *Iface
		peers     []*Peer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "generate",
			wantErr: false,
			fields: fields{
				Interface: &Iface{
					PrivateKey: "oGMS0W0G4JJVNQsUDfdd+lCbXgrqb1DyvGOHrE77gXU=",
					Address:    "10.66.0.1/24",
					ListenPort: 51820,
					DNS:        "8.8.8.8",
					PostUp:     "touch /tmp/wg0.postup",
					PreUp:      "touch /tmp/wg0.preup",
					PreDown:    "touch /tmp/wg0.predown",
					PostDown:   "touch /tmp/wg0.postdown",
				},
				peers: []*Peer{
					{
						Name:       "user1",
						PublicKey:  "gD/gi2mxQ8B6XkmqyFJrpxFkZMjrm7NTBQu3qoFzskA=",
						AllowedIPs: "10.66.0.2/32",
					},
					{
						Name:       "user2",
						PublicKey:  "TrMvSoP4jYQlY6RIzBgbssQqY3vxI2Pi+y71lOWWXX0=",
						AllowedIPs: "10.66.0.3/32",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WG{
				Name:      "wg0",
				CfgDir:    "/tmp",
				Interface: tt.fields.Interface,
				Peers:     tt.fields.peers,
			}
			err := w.Up()
			if (err != nil) != tt.wantErr {
				t.Errorf("WG.UP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = w.Down()
			if (err != nil) != tt.wantErr {
				t.Errorf("WG.Down() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestWG_GenerateClientConf(t *testing.T) {
	type fields struct {
		Name       string
		CfgDir     string
		ExternalIP string
		Interface  *Iface
		Peers      []*Peer
	}
	type args struct {
		peer *Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "generate",
			fields: fields{
				Name:       "wg0",
				CfgDir:     "/tmp",
				ExternalIP: "192.168.0.104",
				Interface: &Iface{
					ListenPort: 51820,
					PublicKey:  "0AzCvh7E6k687t4ZXutKQGC3VCYToJ9yvb6NMYrd6Ts=",
				},
			},
			args: args{
				peer: &Peer{
					PersistentKeepalive: 15,
					PrivateKey:          "6EfkFcsDxSkeDqFJ5mJxgQbayilNN/uzkCaUtf7KEm0=",
					Address:             "10.66.0.2/24",
					AllowedIPs:          "10.66.0.0/24,172.16.0.0/16",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WG{
				Name:       tt.fields.Name,
				CfgDir:     tt.fields.CfgDir,
				ExternalIP: tt.fields.ExternalIP,
				Interface:  tt.fields.Interface,
				Peers:      tt.fields.Peers,
			}
			got, err := w.GenerateClientConf(tt.args.peer)
			if (err != nil) != tt.wantErr {
				t.Errorf("WG.GenerateClientConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
		})
	}
}

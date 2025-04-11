package v1alpha1

type Subnet struct {
	Uuid   string `xorm:"not null unique 'uuid'"`
	Name   string `xorm:"not null unique 'name'"`
	Iface  string `xorm:"not null unique 'iface'"`
	Addr   string `xorm:"not null unique 'addr'"`
	PubKey string `xorm:"not null unique 'pubkey'"`
	PriKey string `xorm:"not null unique 'prikey'"`
}

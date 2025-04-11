package v1alpha1

type Peer struct {
	Uuid   string `xorm:"not null unique 'uuid'"`
	Name   string `xorm:"not null 'name'"`
	Subnet string `xorm:"not null 'subnet'"`
	Addr   string `xorm:"not null unique 'addr'"`
	PubKey string `xorm:"not null unique 'pubkey'"`
	PriKey string `xorm:"not null unique 'prikey'"`
	Enable bool   `xorm:"not null 'enable'"`
}

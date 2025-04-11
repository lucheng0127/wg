package validate

import (
	"github.com/vishvananda/netlink"
)

func ValidatePeerAddr(netAddr, peerAddr string) bool {
	na, err := netlink.ParseAddr(netAddr)
	if err != nil {
		return false
	}

	pa, err := netlink.ParseAddr(peerAddr)
	if err != nil {
		return false
	}

	return na.Contains(pa.IP)
}

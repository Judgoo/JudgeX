package flake

import (
	"github.com/rs/xid"
)

func NextID() string {
	guid := xid.New()
	return guid.String()
}

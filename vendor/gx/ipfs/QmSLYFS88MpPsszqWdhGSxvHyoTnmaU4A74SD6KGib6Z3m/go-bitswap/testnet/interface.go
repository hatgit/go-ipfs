package bitswap

import (
	bsnet "gx/ipfs/QmSLYFS88MpPsszqWdhGSxvHyoTnmaU4A74SD6KGib6Z3m/go-bitswap/network"
	"gx/ipfs/QmcW4FGAt24fdK1jBgWQn3yP4R9ZLyWQqjozv9QK7epRhL/go-testutil"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

type Network interface {
	Adapter(testutil.Identity) bsnet.BitSwapNetwork

	HasPeer(peer.ID) bool
}

package bstest

import (
	. "gx/ipfs/QmNqRBAhovtf4jVd5cF7YvHaFSsQHHZBaUFwGQWPM2CV7R/go-blockservice"

	delay "gx/ipfs/QmRJVNatYJwTAHgdSM1Xef9QVQ1Ch3XHdmcrykjP5Y4soL/go-ipfs-delay"
	bitswap "gx/ipfs/QmSLYFS88MpPsszqWdhGSxvHyoTnmaU4A74SD6KGib6Z3m/go-bitswap"
	tn "gx/ipfs/QmSLYFS88MpPsszqWdhGSxvHyoTnmaU4A74SD6KGib6Z3m/go-bitswap/testnet"
	mockrouting "gx/ipfs/QmbFRJeEmEU16y3BmKKaD4a9fm5oHsEAMHe2vSB1UnfLMi/go-ipfs-routing/mock"
)

// Mocks returns |n| connected mock Blockservices
func Mocks(n int) []BlockService {
	net := tn.VirtualNetwork(mockrouting.NewServer(), delay.Fixed(0))
	sg := bitswap.NewTestSessionGenerator(net)

	instances := sg.Instances(n)

	var servs []BlockService
	for _, i := range instances {
		servs = append(servs, New(i.Blockstore(), i.Exchange))
	}
	return servs
}

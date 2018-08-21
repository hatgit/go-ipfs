package namesys

import (
	"errors"
	"time"

	context "context"

	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	opts "github.com/ipfs/go-ipfs/namesys/opts"
	path "gx/ipfs/QmYKNMEUK7nCVAefgXF1LVtZEZg3uRmBqiae4FJRXDNAyJ/go-path"
	proquint "gx/ipfs/QmYnf27kzqR2cxt6LFZdrAFJuQd6785fTkBvMuEj9EeRxM/proquint"
)

type ProquintResolver struct{}

// Resolve implements Resolver.
func (r *ProquintResolver) Resolve(ctx context.Context, name string, options ...opts.ResolveOpt) (path.Path, error) {
	return resolve(ctx, r, name, opts.ProcessOpts(options), "/ipns/")
}

// resolveOnce implements resolver. Decodes the proquint string.
func (r *ProquintResolver) resolveOnce(ctx context.Context, name string, options *opts.ResolveOpts) (path.Path, time.Duration, error) {
	ok, err := proquint.IsProquint(name)
	if err != nil || !ok {
		return "", 0, errors.New("not a valid proquint string")
	}
	// Return a 0 TTL as caching this result is pointless.
	return path.FromString(string(proquint.Decode(name))), 0, nil
}

func (r *ProquintResolver) secureResolveOnce(ctx context.Context, cw coreiface.ChunkWriter, name string, options *opts.ResolveOpts) (path.Path, time.Duration, error) {
	return r.resolveOnce(ctx, name, options)
}

package routinghelpers

import (
	"context"

	routing "gx/ipfs/QmXijJ3T9MjB2v8xpFDoEX6FqR9u8PkJkzu49TgwJ8Ndr5/go-libp2p-routing"
	ropts "gx/ipfs/QmXijJ3T9MjB2v8xpFDoEX6FqR9u8PkJkzu49TgwJ8Ndr5/go-libp2p-routing/options"

	pstore "gx/ipfs/QmZb7hAgQEhW9dBbzBudU39gCeD4zbe6xafD52LUuF4cUN/go-libp2p-peerstore"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	ci "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	multierror "gx/ipfs/QmfGQp6VVqdPCDyzEM6EGwMY74YPabTSEoQWHUxZuCSWj3/go-multierror"
)

// Tiered is like the Parallel except that GetValue and FindPeer
// are called in series.
type Tiered []routing.IpfsRouting

func (r Tiered) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) error {
	return Parallel(r).PutValue(ctx, key, value, opts...)
}

func (r Tiered) get(ctx context.Context, do func(routing.IpfsRouting) (interface{}, error)) (interface{}, error) {
	var errs []error
	for _, ri := range r {
		val, err := do(ri)
		switch err {
		case nil:
			return val, nil
		case routing.ErrNotFound, routing.ErrNotSupported:
			continue
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		errs = append(errs, err)
	}
	switch len(errs) {
	case 0:
		return nil, routing.ErrNotFound
	case 1:
		return nil, errs[0]
	default:
		return nil, &multierror.Error{Errors: errs}
	}
}

func (r Tiered) GetValue(ctx context.Context, key string, opts ...ropts.Option) ([]byte, error) {
	valInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return ri.GetValue(ctx, key, opts...)
	})
	val, _ := valInt.([]byte)
	return val, err
}

func (r Tiered) GetPublicKey(ctx context.Context, p peer.ID) (ci.PubKey, error) {
	vInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return routing.GetPublicKey(ri, ctx, p)
	})
	val, _ := vInt.(ci.PubKey)
	return val, err
}

func (r Tiered) Provide(ctx context.Context, c *cid.Cid, local bool) error {
	return Parallel(r).Provide(ctx, c, local)
}

func (r Tiered) FindProvidersAsync(ctx context.Context, c *cid.Cid, count int) <-chan pstore.PeerInfo {
	return Parallel(r).FindProvidersAsync(ctx, c, count)
}

func (r Tiered) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	valInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return ri.FindPeer(ctx, p)
	})
	val, _ := valInt.(pstore.PeerInfo)
	return val, err
}

func (r Tiered) Bootstrap(ctx context.Context) error {
	return Parallel(r).Bootstrap(ctx)
}

var _ routing.IpfsRouting = (Tiered)(nil)

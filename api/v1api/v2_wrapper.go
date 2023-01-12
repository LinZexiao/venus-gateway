package v1api

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/filecoin-project/venus/venus-shared/types"
)

type WrapperV2Full struct {
	v2API.IGateway
}

func (w WrapperV2Full) WalletHas(ctx context.Context, account string, addr address.Address) (bool, error) {
	return w.IGateway.WalletHas(ctx, addr, []string{account})
}

func (w WrapperV2Full) WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta types.MsgMeta) (*crypto.Signature, error) {
	return w.IGateway.WalletSign(ctx, addr, []string{account}, toSign, meta)
}

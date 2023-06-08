package marketevent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"

	"github.com/filecoin-project/venus/venus-shared/api"
	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	sharedTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/gateway"

	"github.com/ipfs-force-community/sophon-gateway/types"
)

type MarketEvent struct {
	client        v2API.IMarketServiceProvider
	mAddr         address.Address
	marketHandler types.MarketHandler
	log           *zap.SugaredLogger
	readyCh       chan struct{}
}

func NewMarketRegisterClient(ctx context.Context, url, token string) (v2API.IMarketServiceProvider, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	headers.Add(api.AuthorizationHeader, "Bearer "+token)
	client, closer, err := v2API.NewIGatewayRPC(ctx, url, headers)
	if err != nil {
		return nil, nil, err
	}
	return client, closer, nil
}

func NewMarketEventClient(client v2API.IMarketServiceProvider, mAddr address.Address, marketHandler types.MarketHandler, log *zap.SugaredLogger) *MarketEvent {
	return &MarketEvent{
		client:        client,
		mAddr:         mAddr,
		marketHandler: marketHandler,
		log:           log,
		readyCh:       make(chan struct{}, 1),
	}
}

func (e *MarketEvent) WaitReady(ctx context.Context) {
	select {
	case <-e.readyCh:
	case <-ctx.Done():
	}
}

func (e *MarketEvent) ListenMarketRequest(ctx context.Context) {
	e.log.Infof("start market event listening")
	for {
		if err := e.listenMarketRequestOnce(ctx); err != nil {
			e.log.Errorf("listen market request errored: %s", err)
		} else {
			e.log.Warn("list market request quit")
		}
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			e.log.Warnf("not restarting listen market event: context error: %s", ctx.Err())
			return
		}

		e.log.Info("restarting listen market event ")
		// try clear ready channel
		select {
		case <-e.readyCh:
		default:
		}
	}
}

func (e *MarketEvent) listenMarketRequestOnce(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	policy := &gateway.MarketRegisterPolicy{
		Miner: e.mAddr,
	}
	marketEventCh, err := e.client.ListenMarketEvent(ctx, policy)
	if err != nil {
		// Retry is handled by caller
		return fmt.Errorf("listen market event call failed: %w", err)
	}

	for marketEvent := range marketEventCh {
		switch marketEvent.Method {
		case "InitConnect":
			req := gateway.ConnectedCompleted{}
			err := json.Unmarshal(marketEvent.Payload, &req)
			if err != nil {
				return fmt.Errorf("odd error in connect %v", err)
			}
			e.readyCh <- struct{}{}
			e.log.Infof("success to connect with market %s", req.ChannelId)
		case "SectorsUnsealPiece":
			req := gateway.UnsealRequest{}
			err := json.Unmarshal(marketEvent.Payload, &req)
			if err != nil {
				e.error(ctx, marketEvent.ID, err)
				continue
			}
			err = e.marketHandler.SectorsUnsealPiece(ctx, req.Miner, req.PieceCid, req.Sid, req.Offset, req.Size, req.Dest)
			if err != nil {
				e.error(ctx, marketEvent.ID, err)
				continue
			}
			e.value(ctx, marketEvent.ID, nil)
		default:
			e.log.Errorf("unexpect market event type %s", marketEvent.Method)
		}
	}

	return nil
}

func (e *MarketEvent) value(ctx context.Context, id sharedTypes.UUID, val interface{}) {
	respBytes, err := json.Marshal(val)
	if err != nil {
		e.log.Errorf("marshal address list error %s", err)
		e.error(ctx, id, err)
		return
	}
	err = e.client.ResponseMarketEvent(ctx, &gateway.ResponseEvent{
		ID:      id,
		Payload: respBytes,
		Error:   "",
	})
	if err != nil {
		e.log.Errorf("response error %v", err)
	}
}

func (e *MarketEvent) error(ctx context.Context, id sharedTypes.UUID, err error) {
	err = e.client.ResponseMarketEvent(ctx, &gateway.ResponseEvent{
		ID:      id,
		Payload: nil,
		Error:   err.Error(),
	})
	if err != nil {
		e.log.Errorf("response error %v", err)
	}
}

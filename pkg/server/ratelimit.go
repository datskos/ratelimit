package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/datskos/ratelimit/pkg/proto"
	"github.com/datskos/ratelimit/pkg/storage"
)

type service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) *service {
	return &service{storage}
}

func (service *service) Reduce(context context.Context, request *proto.ReduceRequest) (*proto.ReduceResponse, error) {
	params := newParams(request)
	isValid := params.isValid()

	log.Printf("[request] key=%s max=%d refillAmount=%d refillDuration=%s, isValid=%t",
		params.key, params.maxAmount, params.refillAmount, params.refillDuration, isValid)
	if !isValid {
		return &proto.ReduceResponse{Status: proto.ReduceResponse_ERROR}, nil
	}

	encodedKey := params.encodedKey()
	tx := service.storage.Tx()
	defer tx.Commit() // ok to call twice; necessary for cleaning up in case of failure before normal commit
	value, err := tx.Get(encodedKey)
	if err != nil {
		return nil, err
	}

	if value == nil {
		value = params.newValue()
	} else {
		params.adjustTokens(value)
	}

	status := params.reduce(value)
	err = tx.Set(encodedKey, value)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &proto.ReduceResponse{
		Status:    status,
		Remaining: value.Remaining,
	}, nil
}

type params struct {
	key            string
	maxAmount      uint32
	refillDuration time.Duration
	refillAmount   uint32
	now            time.Time
}

func newParams(request *proto.ReduceRequest) *params {
	maxAmount := request.GetMaxAmount()
	refillAmount := request.GetRefillAmount()
	if refillAmount == 0 {
		refillAmount = maxAmount
	}

	return &params{
		key:            request.GetKey(),
		maxAmount:      maxAmount,
		refillDuration: time.Duration(request.GetRefillDurationSec()) * time.Second,
		refillAmount:   refillAmount,
		now:            time.Now().UTC(),
	}
}

func (params *params) isValid() bool {
	return !(params.key == "" ||
		params.maxAmount < 1 ||
		params.refillDuration < 1 ||
		params.refillAmount > params.maxAmount)
}

func (params *params) encodedKey() string {
	// encode request params in key so that if user mistakenly sends the same key
	// with different params, they are considered separate
	return fmt.Sprintf("%s.max=%d.rd=%d.ra=%d",
		params.key, params.maxAmount, params.refillDuration/time.Second, params.refillAmount)
}

func (params *params) newValue() *storage.Value {
	return &storage.Value{
		Remaining:      params.maxAmount,
		LastRefilledAt: params.now,
		LastReducedAt:  params.now,
	}
}

// mutates Value
func (params *params) adjustTokens(value *storage.Value) {
	elapsed := params.now.Sub(value.LastRefilledAt)
	spans := uint32(elapsed / params.refillDuration)

	value.LastRefilledAt = value.LastRefilledAt.Add(time.Duration(spans) * params.refillDuration)
	value.Remaining = min(params.maxAmount, spans*params.refillAmount+value.Remaining)
}

// mutates Value
func (params *params) reduce(value *storage.Value) proto.ReduceResponse_Status {
	var status proto.ReduceResponse_Status
	if value.Remaining == 0 {
		status = proto.ReduceResponse_NG
		value.Remaining = 0
	} else {
		status = proto.ReduceResponse_OK
		value.Remaining = value.Remaining - 1
	}

	value.LastReducedAt = params.now
	return status
}

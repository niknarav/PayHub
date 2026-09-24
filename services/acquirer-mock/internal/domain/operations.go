package domain

import (
	"fmt"
	"maps"
	"time"
)

type State int

const (
	StateUnknown State = iota
	StateAuthorized
	StateCaptured
	StateVoided
	StateDeclined
)

func (s State) String() string {
	switch s {
	case StateUnknown:
		return "UNKNOWN"
	case StateAuthorized:
		return "AUTHORIZED"
	case StateCaptured:
		return "CAPTURED"
	case StateVoided:
		return "VOIDED"
	case StateDeclined:
		return "DECLINED"
	default:
		return fmt.Sprintf("State(%d)", int(s))
	}
}

type DeclineCode int

const (
	DeclineNone DeclineCode = iota
	DeclineInsufficientFunds
	DeclineCardBlocked
	DeclineLimitExceeded
)

func (c DeclineCode) String() string {
	switch c {
	case DeclineNone:
		return "NONE"
	case DeclineInsufficientFunds:
		return "INSUFFICIENT_FUNDS"
	case DeclineCardBlocked:
		return "CARD_BLOCKED"
	case DeclineLimitExceeded:
		return "LIMIT_EXCEEDED"
	default:
		return fmt.Sprintf("DeclineCode(%d)", int(c))
	}
}

type Operation struct {
	PaymentID       string
	AcquirerRef     string
	CardToken       string
	Currency        string
	RequestedMinor  int64
	AuthorizedMinor int64
	CapturedMinor   int64
	RefundedMinor   int64
	Refunds         map[string]int64
	State           State
	DeclineCode     DeclineCode
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewOperation(paymentID, acquirerRef, cardToken, currency string,
	amount int64, decline DeclineCode, now time.Time) Operation {
	op := Operation{
		PaymentID:      paymentID,
		AcquirerRef:    acquirerRef,
		CardToken:      cardToken,
		Currency:       currency,
		CreatedAt:      now,
		UpdatedAt:      now,
		Refunds:        make(map[string]int64),
		RequestedMinor: amount,
	}

	if decline == DeclineNone {
		op.State = StateAuthorized
		op.AuthorizedMinor = amount
		op.DeclineCode = DeclineNone
	} else {
		op.State = StateDeclined
		op.DeclineCode = decline
	}

	return op
}

func (o Operation) MatchesAuthorize(cardToken string, amount int64, currency string) bool {
	return o.CardToken == cardToken && o.RequestedMinor == amount && o.Currency == currency
}

func (o *Operation) Capture(amount int64, now time.Time) error {
	if o.State != StateAuthorized && o.State != StateCaptured {
		return ErrInvalidState
	}
	if amount != o.AuthorizedMinor {
		return ErrAmountMismatch
	}
	if o.State == StateCaptured {
		return nil
	}
	o.State = StateCaptured
	o.CapturedMinor = amount
	o.UpdatedAt = now
	return nil
}

func (o *Operation) Void(now time.Time) error {
	switch o.State {
	case StateAuthorized:
		o.State = StateVoided
		o.UpdatedAt = now
		return nil
	case StateVoided:
		return nil
	default:
		return ErrInvalidState
	}
}

func (o *Operation) Refund(refundID string, amount int64, now time.Time) error {
	if o.State != StateCaptured {
		return ErrInvalidState
	}
	if prev, ok := o.Refunds[refundID]; ok {
		if prev != amount {
			return ErrIdempotencyConflict
		}
		return nil
	}
	if o.RefundedMinor+amount > o.CapturedMinor {
		return ErrAmountExceeded
	}
	o.Refunds[refundID] = amount
	o.RefundedMinor += amount
	o.UpdatedAt = now
	return nil
}

func (o Operation) Clone() Operation {
	c := o
	c.Refunds = maps.Clone(o.Refunds)
	return c
}

// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package stripe

import (
	"context"

	"github.com/stripe/stripe-go/v72"
	"github.com/zeebo/errs"

	"storj.io/common/uuid"
	"storj.io/storj/satellite/payments"
)

// creditCards is an implementation of payments.CreditCards.
//
// architecture: Service
type creditCards struct {
	service *Service
}

// List returns a list of credit cards for a given payment account.
func (creditCards *creditCards) List(ctx context.Context, userID uuid.UUID) (cards []payments.CreditCard, err error) {
	defer mon.Task()(&ctx, userID)(&err)

	customerID, err := creditCards.service.db.Customers().GetCustomerID(ctx, userID)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	cusParams := &stripe.CustomerParams{Params: stripe.Params{Context: ctx}}
	customer, err := creditCards.service.stripeClient.Customers().Get(customerID, cusParams)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	cardParams := &stripe.PaymentMethodListParams{
		ListParams: stripe.ListParams{Context: ctx},
		Customer:   &customerID,
		Type:       stripe.String(string(stripe.PaymentMethodTypeCard)),
	}

	paymentMethodsIterator := creditCards.service.stripeClient.PaymentMethods().List(cardParams)
	for paymentMethodsIterator.Next() {
		stripeCard := paymentMethodsIterator.PaymentMethod()

		isDefault := false
		if customer.InvoiceSettings.DefaultPaymentMethod != nil {
			isDefault = customer.InvoiceSettings.DefaultPaymentMethod.ID == stripeCard.ID
		}

		cards = append(cards, payments.CreditCard{
			ID:        stripeCard.ID,
			ExpMonth:  int(stripeCard.Card.ExpMonth),
			ExpYear:   int(stripeCard.Card.ExpYear),
			Brand:     string(stripeCard.Card.Brand),
			Last4:     stripeCard.Card.Last4,
			IsDefault: isDefault,
		})
	}

	if err = paymentMethodsIterator.Err(); err != nil {
		return nil, Error.Wrap(err)
	}

	return cards, nil
}

// Add is used to save new credit card, attach it to payment account and make it default.
func (creditCards *creditCards) Add(ctx context.Context, userID uuid.UUID, cardToken string) (_ payments.CreditCard, err error) {
	defer mon.Task()(&ctx, userID, cardToken)(&err)

	customerID, err := creditCards.service.db.Customers().GetCustomerID(ctx, userID)
	if err != nil {
		return payments.CreditCard{}, payments.ErrAccountNotSetup.Wrap(err)
	}

	cardParams := &stripe.PaymentMethodParams{
		Params: stripe.Params{Context: ctx},
		Type:   stripe.String(string(stripe.PaymentMethodTypeCard)),
		Card:   &stripe.PaymentMethodCardParams{Token: &cardToken},
	}

	card, err := creditCards.service.stripeClient.PaymentMethods().New(cardParams)
	if err != nil {
		return payments.CreditCard{}, Error.Wrap(err)
	}

	attachParams := &stripe.PaymentMethodAttachParams{
		Params:   stripe.Params{Context: ctx},
		Customer: &customerID,
	}

	card, err = creditCards.service.stripeClient.PaymentMethods().Attach(card.ID, attachParams)
	if err != nil {
		return payments.CreditCard{}, Error.Wrap(err)
	}

	params := &stripe.CustomerParams{
		Params: stripe.Params{Context: ctx},
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(card.ID),
		},
	}

	_, err = creditCards.service.stripeClient.Customers().Update(customerID, params)

	// TODO: handle created but not attached card manually?
	return payments.CreditCard{
		ID:        card.ID,
		ExpMonth:  int(card.Card.ExpMonth),
		ExpYear:   int(card.Card.ExpYear),
		Brand:     string(card.Card.Brand),
		Last4:     card.Card.Last4,
		IsDefault: true,
	}, Error.Wrap(err)
}

// MakeDefault makes a credit card default payment method.
// this credit card should be attached to account before make it default.
func (creditCards *creditCards) MakeDefault(ctx context.Context, userID uuid.UUID, cardID string) (err error) {
	defer mon.Task()(&ctx, userID, cardID)(&err)

	customerID, err := creditCards.service.db.Customers().GetCustomerID(ctx, userID)
	if err != nil {
		return payments.ErrAccountNotSetup.Wrap(err)
	}

	params := &stripe.CustomerParams{
		Params: stripe.Params{Context: ctx},
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(cardID),
		},
	}

	_, err = creditCards.service.stripeClient.Customers().Update(customerID, params)

	return Error.Wrap(err)
}

// Remove is used to remove credit card from payment account.
func (creditCards *creditCards) Remove(ctx context.Context, userID uuid.UUID, cardID string) (err error) {
	defer mon.Task()(&ctx, cardID)(&err)

	customerID, err := creditCards.service.db.Customers().GetCustomerID(ctx, userID)
	if err != nil {
		return payments.ErrAccountNotSetup.Wrap(err)
	}

	cusParams := &stripe.CustomerParams{Params: stripe.Params{Context: ctx}}
	customer, err := creditCards.service.stripeClient.Customers().Get(customerID, cusParams)
	if err != nil {
		return Error.Wrap(err)
	}
	if customer.InvoiceSettings != nil &&
		customer.InvoiceSettings.DefaultPaymentMethod != nil &&
		customer.InvoiceSettings.DefaultPaymentMethod.ID == cardID {
		return Error.Wrap(errs.New("can not detach default payment method."))
	}

	cardParams := &stripe.PaymentMethodDetachParams{Params: stripe.Params{Context: ctx}}
	_, err = creditCards.service.stripeClient.PaymentMethods().Detach(cardID, cardParams)

	return Error.Wrap(err)
}

// RemoveAll is used to detach all credit cards from payment account.
// It should only be used in case of a user deletion. In case of an error, some cards could have been deleted already.
func (creditCards *creditCards) RemoveAll(ctx context.Context, userID uuid.UUID) (err error) {
	defer mon.Task()(&ctx)(&err)

	ccList, err := creditCards.List(ctx, userID)
	if err != nil {
		return Error.Wrap(err)
	}

	params := &stripe.PaymentMethodDetachParams{Params: stripe.Params{Context: ctx}}
	for _, cc := range ccList {
		_, err = creditCards.service.stripeClient.PaymentMethods().Detach(cc.ID, params)
		if err != nil {
			return Error.Wrap(err)
		}
	}
	return nil
}

package fiat

import (
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// TestGetPrice tests getting price from a set of price data.
func TestGetPrice(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(time.Hour * -1)
	twoHoursAgo := now.Add(time.Hour * -2)

	price10K := decimal.New(10000, 1)
	price20K := decimal.New(20000, 1)

	now10k := &Price{
		Timestamp: now,
		Price:     price10K,
	}

	hourAgo20K := &Price{
		Timestamp: oneHourAgo,
		Price:     price20K,
	}

	tests := []struct {
		name          string
		prices        []*Price
		request       time.Time
		expectedErr   error
		expectedPrice *Price
	}{
		{
			name:        "no prices",
			prices:      nil,
			request:     oneHourAgo,
			expectedErr: errNoPrices,
		},
		{
			name:          "timestamp before range",
			prices:        []*Price{now10k},
			request:       oneHourAgo,
			expectedErr:   errPriceOutOfRange,
			expectedPrice: nil,
		},
		{
			name:          "timestamp equals data point timestamp",
			prices:        []*Price{hourAgo20K, now10k},
			request:       now,
			expectedErr:   nil,
			expectedPrice: now10k,
		},
		{
			name: "timestamp after range",
			prices: []*Price{
				{
					Timestamp: twoHoursAgo,
					Price:     price10K,
				},
				hourAgo20K,
			},
			request:       now,
			expectedErr:   nil,
			expectedPrice: hourAgo20K,
		},
		{
			name:          "timestamp between prices, pick earlier",
			prices:        []*Price{hourAgo20K, now10k},
			request:       now.Add(time.Minute * -30),
			expectedErr:   nil,
			expectedPrice: hourAgo20K,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			price, err := GetPrice(test.prices, test.request)
			if err != test.expectedErr {
				t.Fatalf("expected: %v, got: %v",
					test.expectedErr, err)
			}

			require.Equal(t, test.expectedPrice, price)
		})
	}
}

// TestMSatToUsd tests conversion of msat to usd. This
func TestMSatToUsd(t *testing.T) {
	tests := []struct {
		name         string
		amount       lnwire.MilliSatoshi
		price        decimal.Decimal
		expectedFiat decimal.Decimal
	}{
		{
			name:         "1 sat not rounded down",
			amount:       1000,
			price:        decimal.NewFromInt(10000),
			expectedFiat: decimal.NewFromFloat(0.0001),
		},
		{
			name:         "1 msat not rounded down",
			amount:       1,
			price:        decimal.NewFromInt(10000),
			expectedFiat: decimal.NewFromFloat(0.0000001),
		},
		{
			name:         "1 btc + 1 msat not rounded down",
			amount:       100000000001,
			price:        decimal.NewFromInt(10000),
			expectedFiat: decimal.NewFromFloat(10000.0000001),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			amt := MsatToUSD(test.price, test.amount)
			if !amt.Equals(test.expectedFiat) {
				t.Fatalf("expected: %v, got: %v",
					test.expectedFiat, amt)
			}
		})
	}
}

package exchangerates

import (
	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

// ExchangeRatesDataProvider defines the structure of exchange rates data provider
type ExchangeRatesDataProvider interface {
	// GetLatestExchangeRates returns the common response entities
	GetLatestExchangeRates(c core.Context, uid int64, currentConfig *settings.Config) (*models.LatestExchangeRateResponse, error)
}

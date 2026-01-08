package money

import (
	"errors"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

// ---------- Errors ----------

var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrNoProvider       = errors.New("no exchange rate provider configured")
	ErrRateNotFound     = errors.New("exchange rate not found")
	ErrInvalidFormat    = errors.New("invalid money format")
)

// ---------- Core Types ----------

// Money represents a monetary value with currency.
// Amount is stored as string to preserve decimal precision.
type Money struct {
	Amount   MoneyAmount `json:"amount"`
	Currency Currency    `json:"currency"`
	dec      *decimal.Decimal
}

// MoneyAmount is a string-based amount for precision.
type MoneyAmount string

// Currency represents a currency code.
type Currency string

// Supported currencies with different precision.
const (
	USD Currency = "USD" // Precision: 2 (cents)
	EUR Currency = "EUR" // Precision: 2
	RUB Currency = "RUB" // Precision: 2
	BTC Currency = "BTC" // Precision: 8 (satoshi)
	ETH Currency = "ETH" // Precision: 18 (wei)
)

// ---------- Currency Methods ----------

// Precision returns the number of decimal places for the currency.
func (c Currency) Precision() int32 {
	switch c {
	case BTC:
		return 8
	case ETH:
		return 18
	default:
		return 2
	}
}

// Symbol returns the currency symbol.
func (c Currency) Symbol() string {
	switch c {
	case USD:
		return "$"
	case EUR:
		return "€"
	case RUB:
		return "₽"
	case BTC:
		return "₿"
	case ETH:
		return "Ξ"
	default:
		return string(c)
	}
}

// ---------- Constructors ----------

// New creates a new Money from string amount.
func New(amount string, currency Currency) *Money {
	return &Money{
		Amount:   MoneyAmount(amount),
		Currency: currency,
	}
}

// NewFromSmallestUnit creates Money from smallest unit (cents, satoshi, wei).
func NewFromSmallestUnit(units int64, currency Currency) *Money {
	precision := currency.Precision()
	divisor := decimal.NewFromInt(1).Shift(precision)
	amount := decimal.NewFromInt(units).Div(divisor)
	return &Money{
		Amount:   MoneyAmount(amount.StringFixed(precision)),
		Currency: currency,
	}
}

// NewFromDecimal creates Money from decimal.Decimal.
func NewFromDecimal(d decimal.Decimal, currency Currency) *Money {
	return &Money{
		Amount:   MoneyAmount(d.StringFixed(currency.Precision())),
		Currency: currency,
		dec:      &d,
	}
}

// Zero returns zero Money for the given currency.
func Zero(currency Currency) *Money {
	return New("0", currency)
}

// MustParse parses "100.50 USD" format, panics on error.
func MustParse(s string) *Money {
	m, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return m
}

// Parse parses "100.50 USD" format.
func Parse(s string) (*Money, error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return nil, ErrInvalidFormat
	}

	amount := parts[0]
	currency := Currency(strings.ToUpper(parts[1]))

	m := New(amount, currency)
	if !m.IsValid() {
		return nil, ErrInvalidFormat
	}

	return m, nil
}

// ---------- Decimal Caching ----------

func (m *Money) ensureDecimal() {
	if m.dec == nil {
		d, _ := decimal.NewFromString(string(m.Amount))
		m.dec = &d
	}
}

func (m *Money) decimal() decimal.Decimal {
	m.ensureDecimal()
	return *m.dec
}

// ---------- Arithmetic (Immutable) ----------

// Add adds two Money values. Returns error if currencies don't match.
func (m *Money) Add(other *Money) (*Money, error) {
	if m.Currency != other.Currency {
		return nil, ErrCurrencyMismatch
	}

	result := m.decimal().Add(other.decimal())
	return NewFromDecimal(result, m.Currency), nil
}

// Sub subtracts Money. Returns error if currencies don't match.
func (m *Money) Sub(other *Money) (*Money, error) {
	if m.Currency != other.Currency {
		return nil, ErrCurrencyMismatch
	}

	result := m.decimal().Sub(other.decimal())
	return NewFromDecimal(result, m.Currency), nil
}

// Mul multiplies by a float64 (e.g., tax rate, discount).
func (m *Money) Mul(multiplier float64) *Money {
	result := m.decimal().Mul(decimal.NewFromFloat(multiplier))
	return NewFromDecimal(result, m.Currency)
}

// Div divides by a float64.
func (m *Money) Div(divisor float64) *Money {
	if divisor == 0 {
		return m
	}
	result := m.decimal().Div(decimal.NewFromFloat(divisor))
	return NewFromDecimal(result, m.Currency)
}

// Abs returns absolute value.
func (m *Money) Abs() *Money {
	result := m.decimal().Abs()
	return NewFromDecimal(result, m.Currency)
}

// Neg returns negated value.
func (m *Money) Neg() *Money {
	result := m.decimal().Neg()
	return NewFromDecimal(result, m.Currency)
}

// ---------- Comparison ----------

// Eq returns true if Money values are equal (same currency and amount).
func (m *Money) Eq(other *Money) bool {
	if m == nil || other == nil {
		return m == other
	}
	if m.Currency != other.Currency {
		return false
	}
	return m.decimal().Equal(other.decimal())
}

// Gt returns true if m > other.
func (m *Money) Gt(other *Money) bool {
	return m.decimal().GreaterThan(other.decimal())
}

// Gte returns true if m >= other.
func (m *Money) Gte(other *Money) bool {
	return m.decimal().GreaterThanOrEqual(other.decimal())
}

// Lt returns true if m < other.
func (m *Money) Lt(other *Money) bool {
	return m.decimal().LessThan(other.decimal())
}

// Lte returns true if m <= other.
func (m *Money) Lte(other *Money) bool {
	return m.decimal().LessThanOrEqual(other.decimal())
}

// IsZero returns true if amount is zero.
func (m *Money) IsZero() bool {
	return m.decimal().IsZero()
}

// IsPositive returns true if amount > 0.
func (m *Money) IsPositive() bool {
	return m.decimal().IsPositive()
}

// IsNegative returns true if amount < 0.
func (m *Money) IsNegative() bool {
	return m.decimal().IsNegative()
}

// ---------- Conversion ----------

// ToSmallestUnit returns amount in smallest unit (cents, satoshi, wei).
func (m *Money) ToSmallestUnit() int64 {
	multiplier := decimal.NewFromInt(1).Shift(m.Currency.Precision())
	return m.decimal().Mul(multiplier).IntPart()
}

// String returns "100.50 USD" format.
func (m *Money) String() string {
	return m.decimal().StringFixed(m.Currency.Precision()) + " " + string(m.Currency)
}

// StringAmount returns just the amount "100.50".
func (m *Money) StringAmount() string {
	return m.decimal().StringFixed(m.Currency.Precision())
}

// StringFormatted returns formatted with symbol "$100.50".
func (m *Money) StringFormatted() string {
	return m.Currency.Symbol() + m.StringAmount()
}

// ---------- Exchange Rates ----------

// ExchangeRateProvider provides exchange rates between currencies.
type ExchangeRateProvider interface {
	GetRate(from, to Currency) (float64, error)
}

var defaultProvider ExchangeRateProvider

// SetDefaultProvider sets the default exchange rate provider.
// Call this once at application startup.
func SetDefaultProvider(p ExchangeRateProvider) {
	defaultProvider = p
}

// DefaultProvider returns the current default provider.
func DefaultProvider() ExchangeRateProvider {
	return defaultProvider
}

// ConvertTo converts to another currency using the default provider.
func (m *Money) ConvertTo(currency Currency) (*Money, error) {
	if defaultProvider == nil {
		return nil, ErrNoProvider
	}
	return m.ConvertToWith(currency, defaultProvider)
}

// ConvertToWith converts using an explicit provider (for testing).
func (m *Money) ConvertToWith(currency Currency, provider ExchangeRateProvider) (*Money, error) {
	if m.Currency == currency {
		return m, nil
	}

	rate, err := provider.GetRate(m.Currency, currency)
	if err != nil {
		return nil, err
	}

	result := m.decimal().Mul(decimal.NewFromFloat(rate))
	return NewFromDecimal(result, currency), nil
}

// ---------- Static Rate Provider ----------

// StaticRateProvider provides static exchange rates (useful for testing).
type StaticRateProvider struct {
	Rates map[Currency]map[Currency]float64
}

// NewStaticProvider creates a provider with static rates.
func NewStaticProvider(rates map[Currency]map[Currency]float64) *StaticRateProvider {
	return &StaticRateProvider{Rates: rates}
}

// GetRate returns the exchange rate from one currency to another.
func (p *StaticRateProvider) GetRate(from, to Currency) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	fromRates, ok := p.Rates[from]
	if !ok {
		return 0, ErrRateNotFound
	}

	rate, ok := fromRates[to]
	if !ok {
		return 0, ErrRateNotFound
	}

	return rate, nil
}

// ---------- Validation ----------

var amountRegex = regexp.MustCompile(`^-?\d{1,15}(\.\d{1,18})?$`)

// IsValid returns true if Money has valid amount and currency.
func (m *Money) IsValid() bool {
	if m == nil {
		return false
	}
	if m.Currency == "" {
		return false
	}
	return amountRegex.MatchString(string(m.Amount))
}

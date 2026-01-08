package money_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myapp/internal/money"
)

// ---------- Precision Tests ----------

func TestFloatPrecisionProblem(t *testing.T) {
	// Demonstrates why we don't use float64 for money
	a := 0.1
	b := 0.2
	sum := a + b

	// Float has precision issues
	assert.NotEqual(t, 0.3, sum, "float64 has precision issues: 0.1 + 0.2 != 0.3")

	// Money type handles this correctly
	moneyA := money.New("0.1", money.USD)
	moneyB := money.New("0.2", money.USD)
	moneySum, _ := moneyA.Add(moneyB)

	expected := money.New("0.3", money.USD)
	assert.True(t, moneySum.Eq(expected), "Money type: 0.1 + 0.2 = 0.3")
}

// ---------- Constructor Tests ----------

func TestNew(t *testing.T) {
	m := money.New("100.50", money.USD)

	assert.Equal(t, money.MoneyAmount("100.50"), m.Amount)
	assert.Equal(t, money.USD, m.Currency)
}

func TestNewFromSmallestUnit(t *testing.T) {
	tests := []struct {
		name     string
		units    int64
		currency money.Currency
		expected string
	}{
		{"USD cents", 10050, money.USD, "100.50"},
		{"USD zero", 0, money.USD, "0.00"},
		{"BTC satoshi", 100000000, money.BTC, "1.00000000"},
		{"BTC small", 1, money.BTC, "0.00000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := money.NewFromSmallestUnit(tt.units, tt.currency)
			assert.Equal(t, tt.expected, m.StringAmount())
		})
	}
}

func TestZero(t *testing.T) {
	m := money.Zero(money.USD)

	assert.True(t, m.IsZero())
	assert.Equal(t, "0.00", m.StringAmount())
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		amount  string
		curr    money.Currency
	}{
		{"valid USD", "100.50 USD", false, "100.50", money.USD},
		{"valid EUR", "50 EUR", false, "50", money.EUR},
		{"lowercase", "100 usd", false, "100", money.USD},
		{"invalid format", "invalid", true, "", ""},
		{"no currency", "100.50", true, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := money.Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.curr, m.Currency)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	m := money.MustParse("100.50 USD")
	assert.Equal(t, money.USD, m.Currency)

	assert.Panics(t, func() {
		money.MustParse("invalid")
	})
}

// ---------- Arithmetic Tests ----------

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a        *money.Money
		b        *money.Money
		expected string
		wantErr  bool
	}{
		{
			name:     "simple add",
			a:        money.New("100.00", money.USD),
			b:        money.New("50.50", money.USD),
			expected: "150.50",
		},
		{
			name:     "add with decimals",
			a:        money.New("0.10", money.USD),
			b:        money.New("0.20", money.USD),
			expected: "0.30",
		},
		{
			name:    "currency mismatch",
			a:       money.New("100", money.USD),
			b:       money.New("100", money.EUR),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.a.Add(tt.b)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, money.ErrCurrencyMismatch)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.StringAmount())
			}
		})
	}
}

func TestSub(t *testing.T) {
	a := money.New("100.00", money.USD)
	b := money.New("30.50", money.USD)

	result, err := a.Sub(b)
	require.NoError(t, err)
	assert.Equal(t, "69.50", result.StringAmount())

	// Negative result
	result, err = b.Sub(a)
	require.NoError(t, err)
	assert.Equal(t, "-69.50", result.StringAmount())
	assert.True(t, result.IsNegative())
}

func TestMul(t *testing.T) {
	m := money.New("100.00", money.USD)

	// 15% tax
	tax := m.Mul(0.15)
	assert.Equal(t, "15.00", tax.StringAmount())

	// 50% discount
	discount := m.Mul(0.5)
	assert.Equal(t, "50.00", discount.StringAmount())
}

func TestDiv(t *testing.T) {
	m := money.New("100.00", money.USD)

	// Split between 3 people
	split := m.Div(3)
	assert.Equal(t, "33.33", split.StringAmount())

	// Division by zero returns original
	same := m.Div(0)
	assert.True(t, same.Eq(m))
}

func TestAbs(t *testing.T) {
	negative := money.New("-50.00", money.USD)
	positive := negative.Abs()

	assert.Equal(t, "50.00", positive.StringAmount())
	assert.True(t, positive.IsPositive())
}

func TestNeg(t *testing.T) {
	positive := money.New("50.00", money.USD)
	negative := positive.Neg()

	assert.Equal(t, "-50.00", negative.StringAmount())
	assert.True(t, negative.IsNegative())
}

// ---------- Comparison Tests ----------

func TestEq(t *testing.T) {
	a := money.New("100.00", money.USD)
	b := money.New("100.00", money.USD)
	c := money.New("100.00", money.EUR)
	d := money.New("50.00", money.USD)

	assert.True(t, a.Eq(b), "same amount and currency")
	assert.False(t, a.Eq(c), "different currency")
	assert.False(t, a.Eq(d), "different amount")
}

func TestComparisons(t *testing.T) {
	small := money.New("50.00", money.USD)
	large := money.New("100.00", money.USD)

	assert.True(t, large.Gt(small))
	assert.True(t, large.Gte(small))
	assert.True(t, small.Lt(large))
	assert.True(t, small.Lte(large))

	same := money.New("100.00", money.USD)
	assert.True(t, large.Gte(same))
	assert.True(t, large.Lte(same))
}

func TestIsZero(t *testing.T) {
	zero := money.Zero(money.USD)
	nonZero := money.New("0.01", money.USD)

	assert.True(t, zero.IsZero())
	assert.False(t, nonZero.IsZero())
}

func TestIsPositiveNegative(t *testing.T) {
	positive := money.New("100.00", money.USD)
	negative := money.New("-100.00", money.USD)
	zero := money.Zero(money.USD)

	assert.True(t, positive.IsPositive())
	assert.False(t, positive.IsNegative())

	assert.True(t, negative.IsNegative())
	assert.False(t, negative.IsPositive())

	assert.False(t, zero.IsPositive())
	assert.False(t, zero.IsNegative())
}

// ---------- Conversion Tests ----------

func TestToSmallestUnit(t *testing.T) {
	tests := []struct {
		name     string
		money    *money.Money
		expected int64
	}{
		{"USD dollars", money.New("100.50", money.USD), 10050},
		{"USD cents", money.New("0.01", money.USD), 1},
		{"BTC satoshi", money.New("1.00000001", money.BTC), 100000001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.money.ToSmallestUnit())
		})
	}
}

func TestString(t *testing.T) {
	m := money.New("100.50", money.USD)

	assert.Equal(t, "100.50 USD", m.String())
	assert.Equal(t, "100.50", m.StringAmount())
	assert.Equal(t, "$100.50", m.StringFormatted())
}

// ---------- Exchange Rate Tests ----------

func TestStaticRateProvider(t *testing.T) {
	provider := money.NewStaticProvider(map[money.Currency]map[money.Currency]float64{
		money.USD: {money.EUR: 0.85, money.BTC: 0.000024},
		money.EUR: {money.USD: 1.18},
	})

	// USD to EUR
	rate, err := provider.GetRate(money.USD, money.EUR)
	require.NoError(t, err)
	assert.Equal(t, 0.85, rate)

	// Same currency
	rate, err = provider.GetRate(money.USD, money.USD)
	require.NoError(t, err)
	assert.Equal(t, 1.0, rate)

	// Rate not found
	_, err = provider.GetRate(money.EUR, money.BTC)
	assert.ErrorIs(t, err, money.ErrRateNotFound)
}

func TestConvertToWith(t *testing.T) {
	provider := money.NewStaticProvider(map[money.Currency]map[money.Currency]float64{
		money.USD: {money.EUR: 0.85},
	})

	usd := money.New("100.00", money.USD)
	eur, err := usd.ConvertToWith(money.EUR, provider)

	require.NoError(t, err)
	assert.Equal(t, money.EUR, eur.Currency)
	assert.Equal(t, "85.00", eur.StringAmount())
}

func TestConvertTo_NoProvider(t *testing.T) {
	// Ensure no default provider is set
	money.SetDefaultProvider(nil)

	usd := money.New("100.00", money.USD)
	_, err := usd.ConvertTo(money.EUR)

	assert.ErrorIs(t, err, money.ErrNoProvider)
}

func TestConvertTo_WithDefaultProvider(t *testing.T) {
	provider := money.NewStaticProvider(map[money.Currency]map[money.Currency]float64{
		money.USD: {money.EUR: 0.85},
	})
	money.SetDefaultProvider(provider)
	defer money.SetDefaultProvider(nil)

	usd := money.New("100.00", money.USD)
	eur, err := usd.ConvertTo(money.EUR)

	require.NoError(t, err)
	assert.Equal(t, "85.00", eur.StringAmount())
}

func TestConvertTo_SameCurrency(t *testing.T) {
	usd := money.New("100.00", money.USD)
	result, err := usd.ConvertToWith(money.USD, nil)

	require.NoError(t, err)
	assert.True(t, result.Eq(usd))
}

// ---------- Validation Tests ----------

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		money *money.Money
		valid bool
	}{
		{"valid", money.New("100.50", money.USD), true},
		{"zero", money.Zero(money.USD), true},
		{"negative", money.New("-50.00", money.USD), true},
		{"no currency", money.New("100", ""), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.money.IsValid())
		})
	}
}

// ---------- Currency Tests ----------

func TestCurrencyPrecision(t *testing.T) {
	assert.Equal(t, int32(2), money.USD.Precision())
	assert.Equal(t, int32(2), money.EUR.Precision())
	assert.Equal(t, int32(8), money.BTC.Precision())
	assert.Equal(t, int32(18), money.ETH.Precision())
}

func TestCurrencySymbol(t *testing.T) {
	assert.Equal(t, "$", money.USD.Symbol())
	assert.Equal(t, "€", money.EUR.Symbol())
	assert.Equal(t, "₿", money.BTC.Symbol())
	assert.Equal(t, "Ξ", money.ETH.Symbol())
}

// ---------- Edge Cases ----------

func TestHighPrecisionCrypto(t *testing.T) {
	// BTC with 8 decimal places
	btc := money.New("0.00000001", money.BTC) // 1 satoshi
	assert.Equal(t, "0.00000001", btc.StringAmount())
	assert.Equal(t, int64(1), btc.ToSmallestUnit())

	// Addition of small BTC amounts
	a := money.New("0.00000001", money.BTC)
	b := money.New("0.00000002", money.BTC)
	sum, _ := a.Add(b)
	assert.Equal(t, "0.00000003", sum.StringAmount())
}

func TestLargeAmounts(t *testing.T) {
	// Large amount
	large := money.New("999999999999.99", money.USD)
	assert.True(t, large.IsValid())
	assert.Equal(t, "999999999999.99", large.StringAmount())

	// Add large amounts
	sum, _ := large.Add(large)
	assert.Equal(t, "1999999999999.98", sum.StringAmount())
}

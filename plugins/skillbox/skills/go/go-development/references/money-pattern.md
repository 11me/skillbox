# Money Pattern

Safe money handling with decimal precision and currency conversion.

## Problem

Float64 has precision issues in financial calculations:

```go
// Float64 precision problem
a := 0.1
b := 0.2
fmt.Println(a + b) // 0.30000000000000004 ← WRONG!

// Accumulating errors
sum := 0.0
for i := 0; i < 10; i++ {
    sum += 0.1
}
fmt.Println(sum) // 0.9999999999999999 ← NOT 1.0!
```

**Why this happens:** Float64 uses binary representation, which cannot exactly represent decimal fractions like 0.1.

## Solution

String-based amount storage with `shopspring/decimal` for arithmetic:

```go
type Money struct {
    Amount   MoneyAmount `json:"amount"`
    Currency Currency    `json:"currency"`
    dec      *decimal.Decimal // cached for performance
}

type MoneyAmount string  // "100.50" — string preserves precision
type Currency string     // "USD", "EUR", "BTC"
```

**Key design decisions:**
- **String storage** — preserves exact decimal representation
- **Cached decimal** — computed once, reused for operations
- **Immutable operations** — methods return new Money, never mutate

## Currency with Precision

Different currencies have different precision requirements:

```go
type Currency string

const (
    USD Currency = "USD"  // 2 decimals (cents)
    EUR Currency = "EUR"  // 2 decimals
    BTC Currency = "BTC"  // 8 decimals (satoshi)
    ETH Currency = "ETH"  // 18 decimals (wei)
)

func (c Currency) Precision() int32 {
    switch c {
    case BTC:
        return 8
    case ETH:
        return 18
    default:
        return 2  // Fiat default
    }
}

func (c Currency) Symbol() string {
    switch c {
    case USD:
        return "$"
    case EUR:
        return "€"
    case BTC:
        return "₿"
    case ETH:
        return "Ξ"
    default:
        return string(c)
    }
}
```

## Constructors

```go
// From string amount
func New(amount string, currency Currency) *Money {
    return &Money{
        Amount:   MoneyAmount(amount),
        Currency: currency,
    }
}

// From smallest unit (cents, satoshi, wei)
func NewFromSmallestUnit(units int64, currency Currency) *Money {
    precision := currency.Precision()
    divisor := decimal.NewFromInt(1).Shift(precision)
    amount := decimal.NewFromInt(units).Div(divisor)
    return &Money{
        Amount:   MoneyAmount(amount.StringFixed(precision)),
        Currency: currency,
    }
}

// Zero value
func Zero(currency Currency) *Money {
    return New("0", currency)
}

// Parse from string "100.50 USD"
func MustParse(s string) *Money {
    m, err := Parse(s)
    if err != nil {
        panic(err)
    }
    return m
}
```

## Arithmetic Operations

All operations return new Money (immutable):

```go
// Add — requires same currency
func (m *Money) Add(other *Money) (*Money, error) {
    if m.Currency != other.Currency {
        return nil, ErrCurrencyMismatch
    }
    m.ensureDecimal()
    other.ensureDecimal()

    result := m.dec.Add(*other.dec)
    return &Money{
        Amount:   MoneyAmount(result.StringFixed(m.Currency.Precision())),
        Currency: m.Currency,
        dec:      &result,
    }, nil
}

// Sub — requires same currency
func (m *Money) Sub(other *Money) (*Money, error) {
    if m.Currency != other.Currency {
        return nil, ErrCurrencyMismatch
    }
    m.ensureDecimal()
    other.ensureDecimal()

    result := m.dec.Sub(*other.dec)
    return &Money{
        Amount:   MoneyAmount(result.StringFixed(m.Currency.Precision())),
        Currency: m.Currency,
        dec:      &result,
    }, nil
}

// Mul — multiply by float (e.g., tax rate, discount)
func (m *Money) Mul(multiplier float64) *Money {
    m.ensureDecimal()
    result := m.dec.Mul(decimal.NewFromFloat(multiplier))
    return &Money{
        Amount:   MoneyAmount(result.StringFixed(m.Currency.Precision())),
        Currency: m.Currency,
        dec:      &result,
    }
}

// Div — divide by float
func (m *Money) Div(divisor float64) *Money {
    m.ensureDecimal()
    result := m.dec.Div(decimal.NewFromFloat(divisor))
    return &Money{
        Amount:   MoneyAmount(result.StringFixed(m.Currency.Precision())),
        Currency: m.Currency,
        dec:      &result,
    }
}

// Abs — absolute value
func (m *Money) Abs() *Money {
    m.ensureDecimal()
    result := m.dec.Abs()
    return &Money{
        Amount:   MoneyAmount(result.StringFixed(m.Currency.Precision())),
        Currency: m.Currency,
        dec:      &result,
    }
}

// Neg — negate
func (m *Money) Neg() *Money {
    m.ensureDecimal()
    result := m.dec.Neg()
    return &Money{
        Amount:   MoneyAmount(result.StringFixed(m.Currency.Precision())),
        Currency: m.Currency,
        dec:      &result,
    }
}

// Helper to cache decimal
func (m *Money) ensureDecimal() {
    if m.dec == nil {
        d, _ := decimal.NewFromString(string(m.Amount))
        m.dec = &d
    }
}
```

## Comparison Operations

```go
func (m *Money) Eq(other *Money) bool {
    if m.Currency != other.Currency {
        return false
    }
    m.ensureDecimal()
    other.ensureDecimal()
    return m.dec.Equal(*other.dec)
}

func (m *Money) Gt(other *Money) bool {
    m.ensureDecimal()
    other.ensureDecimal()
    return m.dec.GreaterThan(*other.dec)
}

func (m *Money) Gte(other *Money) bool {
    m.ensureDecimal()
    other.ensureDecimal()
    return m.dec.GreaterThanOrEqual(*other.dec)
}

func (m *Money) Lt(other *Money) bool {
    m.ensureDecimal()
    other.ensureDecimal()
    return m.dec.LessThan(*other.dec)
}

func (m *Money) Lte(other *Money) bool {
    m.ensureDecimal()
    other.ensureDecimal()
    return m.dec.LessThanOrEqual(*other.dec)
}

func (m *Money) IsZero() bool {
    m.ensureDecimal()
    return m.dec.IsZero()
}

func (m *Money) IsPositive() bool {
    m.ensureDecimal()
    return m.dec.IsPositive()
}

func (m *Money) IsNegative() bool {
    m.ensureDecimal()
    return m.dec.IsNegative()
}
```

## Currency Conversion

### Exchange Rate Provider Interface

```go
type ExchangeRateProvider interface {
    GetRate(from, to Currency) (float64, error)
}

var (
    defaultProvider ExchangeRateProvider
    ErrNoProvider   = errors.New("no exchange rate provider configured")
    ErrRateNotFound = errors.New("exchange rate not found")
)

// Set default provider at startup
func SetDefaultProvider(p ExchangeRateProvider) {
    defaultProvider = p
}

func DefaultProvider() ExchangeRateProvider {
    return defaultProvider
}
```

### Conversion Methods

```go
// ConvertTo — uses default provider (clean API)
func (m *Money) ConvertTo(currency Currency) (*Money, error) {
    if defaultProvider == nil {
        return nil, ErrNoProvider
    }
    return m.ConvertToWith(currency, defaultProvider)
}

// ConvertToWith — explicit provider (for tests)
func (m *Money) ConvertToWith(currency Currency, provider ExchangeRateProvider) (*Money, error) {
    if m.Currency == currency {
        return m, nil
    }

    rate, err := provider.GetRate(m.Currency, currency)
    if err != nil {
        return nil, err
    }

    return m.Mul(rate).withCurrency(currency), nil
}

func (m *Money) withCurrency(c Currency) *Money {
    m.ensureDecimal()
    return &Money{
        Amount:   MoneyAmount(m.dec.StringFixed(c.Precision())),
        Currency: c,
        dec:      m.dec,
    }
}
```

### Static Rate Provider (for testing)

```go
type StaticRateProvider struct {
    Rates map[Currency]map[Currency]float64
}

func NewStaticProvider(rates map[Currency]map[Currency]float64) *StaticRateProvider {
    return &StaticRateProvider{Rates: rates}
}

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

// Usage in tests
func TestConversion(t *testing.T) {
    provider := NewStaticProvider(map[Currency]map[Currency]float64{
        USD: {EUR: 0.85, BTC: 0.000024},
        EUR: {USD: 1.18},
    })

    usd := New("100", USD)
    eur, _ := usd.ConvertToWith(EUR, provider)
    // eur.Amount = "85.00"
}
```

### Usage in Application

```go
// main.go — set provider once at startup
func main() {
    // Use your preferred provider
    money.SetDefaultProvider(rates.NewCoinGeckoProvider())

    // ... rest of app
}

// service.go — clean API without provider argument
func (s *OrderService) CalculateTotal(items []Item) (*money.Money, error) {
    total := money.Zero(money.USD)

    for _, item := range items {
        price := item.Price
        if price.Currency != money.USD {
            var err error
            price, err = price.ConvertTo(money.USD)  // uses default provider
            if err != nil {
                return nil, err
            }
        }
        total, _ = total.Add(price)
    }

    return total, nil
}
```

## Database Storage

Store as two columns: `amount TEXT` + `currency VARCHAR`:

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    total_amount TEXT NOT NULL,      -- "99.99"
    total_currency VARCHAR(10) NOT NULL,  -- "USD"
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Repository Pattern

```go
type orderMapper struct {
    ID            uuid.UUID
    TotalAmount   string  // scanned from TEXT
    TotalCurrency string  // scanned from VARCHAR
    CreatedAt     time.Time
}

func (m *orderMapper) toDomain() *Order {
    return &Order{
        ID:        m.ID,
        Total:     money.New(m.TotalAmount, money.Currency(m.TotalCurrency)),
        CreatedAt: m.CreatedAt,
    }
}

// Insert
func (r *OrderRepository) Create(ctx context.Context, order *Order) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO orders (id, total_amount, total_currency)
        VALUES ($1, $2, $3)
    `, order.ID, string(order.Total.Amount), string(order.Total.Currency))
    return err
}

// Select with type cast
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*Order, error) {
    row := r.db.QueryRow(ctx, `
        SELECT id, total_amount::text, total_currency, created_at
        FROM orders WHERE id = $1
    `, id)

    var m orderMapper
    err := row.Scan(&m.ID, &m.TotalAmount, &m.TotalCurrency, &m.CreatedAt)
    if err != nil {
        return nil, err
    }
    return m.toDomain(), nil
}
```

## JSON Serialization

Money serializes as JSON object:

```go
// Money JSON representation
{
    "amount": "99.99",
    "currency": "USD"
}
```

Works automatically with struct tags:

```go
type Order struct {
    ID    uuid.UUID `json:"id"`
    Total *Money    `json:"total"`
}

// Serializes to:
// {"id": "...", "total": {"amount": "99.99", "currency": "USD"}}
```

## Validation

```go
var amountRegex = regexp.MustCompile(`^\d{1,15}(\.\d{1,18})?$`)

func (m *Money) IsValid() bool {
    if m == nil {
        return false
    }
    if m.Currency == "" {
        return false
    }
    return amountRegex.MatchString(string(m.Amount))
}

func Parse(s string) (*Money, error) {
    parts := strings.Split(s, " ")
    if len(parts) != 2 {
        return nil, errors.New("invalid format, expected 'amount currency'")
    }

    amount, currency := parts[0], Currency(strings.ToUpper(parts[1]))
    m := New(amount, currency)

    if !m.IsValid() {
        return nil, errors.New("invalid money value")
    }

    return m, nil
}
```

## Best Practices

| DO | DON'T |
|----|-------|
| Use `Money` type for all financial values | Use `float64` for money |
| Store amount as string/TEXT | Store as DECIMAL without precision control |
| Use `Eq()` for comparison | Use `==` operator |
| Check currency before Add/Sub | Assume same currency |
| Set provider at startup | Create provider per request |
| Use `ConvertToWith` in tests | Mock default provider globally |

## Common Pitfalls

### Float Comparison

```go
// ❌ WRONG — float comparison
if order.Total == 99.99 { ... }

// ✅ CORRECT — use Eq()
expected := money.New("99.99", money.USD)
if order.Total.Eq(expected) { ... }
```

### Currency Mismatch

```go
// ❌ WRONG — will return error
usd := money.New("100", money.USD)
eur := money.New("85", money.EUR)
sum, err := usd.Add(eur)  // err: currency mismatch

// ✅ CORRECT — convert first
eurConverted, _ := eur.ConvertTo(money.USD)
sum, _ := usd.Add(eurConverted)
```

### Percentage Calculation

```go
// Calculate 15% discount
price := money.New("100", money.USD)
discount := price.Mul(0.15)      // $15.00
final := price.Sub(discount)      // Error: returns (*Money, error)

// ✅ CORRECT
final, _ := price.Sub(discount)   // $85.00
```

## Dependencies

```bash
go get github.com/shopspring/decimal@latest
```

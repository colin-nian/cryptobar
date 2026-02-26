package price

import (
	"fmt"
	"sync"
)

type PriceData struct {
	Pair      string
	Price     float64
	OpenPrice float64 // 24h open price for calculating change
}

func (p PriceData) Change24h() float64 {
	if p.OpenPrice == 0 {
		return 0
	}
	return (p.Price - p.OpenPrice) / p.OpenPrice * 100
}

func (p PriceData) FormatPrice(precision int) string {
	if precision < 0 {
		precision = autoPrec(p.Price)
	}
	return fmt.Sprintf("%.*f", precision, p.Price)
}

func autoPrec(price float64) int {
	switch {
	case price >= 1000:
		return 0
	case price >= 1:
		return 2
	case price >= 0.01:
		return 4
	default:
		return 6
	}
}

type UpdateCallback func(pair string, data PriceData)

type Store struct {
	mu        sync.RWMutex
	prices    map[string]PriceData
	callbacks []UpdateCallback
}

func NewStore() *Store {
	return &Store{
		prices: make(map[string]PriceData),
	}
}

func (s *Store) Update(pair string, price, openPrice float64) {
	s.mu.Lock()
	data := PriceData{
		Pair:      pair,
		Price:     price,
		OpenPrice: openPrice,
	}
	s.prices[pair] = data
	callbacks := make([]UpdateCallback, len(s.callbacks))
	copy(callbacks, s.callbacks)
	s.mu.Unlock()

	for _, cb := range callbacks {
		cb(pair, data)
	}
}

func (s *Store) Get(pair string) (PriceData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.prices[pair]
	return data, ok
}

func (s *Store) GetAll() map[string]PriceData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]PriceData, len(s.prices))
	for k, v := range s.prices {
		result[k] = v
	}
	return result
}

func (s *Store) OnUpdate(cb UpdateCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks = append(s.callbacks, cb)
}

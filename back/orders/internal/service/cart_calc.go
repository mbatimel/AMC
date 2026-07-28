package service

import "math"

type CartCalcItem struct {
	Qty   int
	Price float64
}

type CartTotals struct {
	Subtotal      float64
	DiscountTotal float64
	VATTotal      float64
	Total         float64
}

func sumLineTotals(items []CartCalcItem) float64 {
	var subtotal float64
	for _, item := range items {
		subtotal += float64(item.Qty) * item.Price
	}
	return round2(subtotal)
}

func calcCartTotals(subtotal float64, discountPercent float64, vatRatePercent float64) CartTotals {
	discountTotal := round2(subtotal * discountPercent / 100)
	afterDiscount := subtotal - discountTotal
	vatTotal := round2(afterDiscount * vatRatePercent / 100)
	total := round2(afterDiscount + vatTotal)

	return CartTotals{
		Subtotal:      round2(subtotal),
		DiscountTotal: discountTotal,
		VATTotal:      vatTotal,
		Total:         total,
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

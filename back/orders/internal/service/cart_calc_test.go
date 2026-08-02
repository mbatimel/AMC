package service

import "testing"

func TestSumLineTotals(t *testing.T) {
	items := []CartCalcItem{
		{Qty: 2, Price: 100},
		{Qty: 3, Price: 50},
	}

	got := sumLineTotals(items)
	want := 350.0
	if got != want {
		t.Fatalf("sumLineTotals() = %v, want %v", got, want)
	}
}

func TestCalcCartTotals_NoDiscountNoVAT(t *testing.T) {
	totals := calcCartTotals(200, 0, 0)
	if totals.Subtotal != 200 || totals.DiscountTotal != 0 || totals.VATTotal != 0 || totals.Total != 200 {
		t.Fatalf("unexpected totals: %+v", totals)
	}
}

func TestCalcCartTotals_DiscountAndVAT(t *testing.T) {
	totals := calcCartTotals(1000, 10, 22)
	if totals.Subtotal != 1000 {
		t.Fatalf("Subtotal = %v, want 1000", totals.Subtotal)
	}
	if totals.DiscountTotal != 100 {
		t.Fatalf("DiscountTotal = %v, want 100", totals.DiscountTotal)
	}
	if totals.VATTotal != 198 {
		t.Fatalf("VATTotal = %v, want 198", totals.VATTotal)
	}
	if totals.Total != 1098 {
		t.Fatalf("Total = %v, want 1098", totals.Total)
	}
}

func TestCalcCartTotals_RoundingToTwoDecimals(t *testing.T) {
	totals := calcCartTotals(99.99, 0, 22)
	if totals.VATTotal != 22.0 {
		t.Fatalf("VATTotal = %v, want 22.0", totals.VATTotal)
	}
	if totals.Total != 121.99 {
		t.Fatalf("Total = %v, want 121.99", totals.Total)
	}
}

func TestCalcEffectiveDiscount_NoPromo(t *testing.T) {
	got := calcEffectiveDiscount(5, 0)
	if got != 5 {
		t.Fatalf("calcEffectiveDiscount(5, 0) = %v, want 5", got)
	}
}

func TestCalcEffectiveDiscount_PromoHigherThanManual(t *testing.T) {
	got := calcEffectiveDiscount(5, 20)
	if got != 20 {
		t.Fatalf("calcEffectiveDiscount(5, 20) = %v, want 20", got)
	}
}

func TestCalcEffectiveDiscount_ManualHigherThanPromo(t *testing.T) {
	got := calcEffectiveDiscount(30, 10)
	if got != 30 {
		t.Fatalf("calcEffectiveDiscount(30, 10) = %v, want 30", got)
	}
}

func TestCalcEffectiveDiscount_Equal(t *testing.T) {
	got := calcEffectiveDiscount(15, 15)
	if got != 15 {
		t.Fatalf("calcEffectiveDiscount(15, 15) = %v, want 15", got)
	}
}

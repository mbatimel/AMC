export const DEFAULT_LOW_STOCK_THRESHOLD = 10;

export type StockLevel = 'in_stock' | 'low' | 'out';

export const getStockLevel = (
  stockQty: number,
  threshold = DEFAULT_LOW_STOCK_THRESHOLD,
): StockLevel => {
  if (stockQty <= 0) {
    return 'out';
  }

  if (stockQty < threshold) {
    return 'low';
  }

  return 'in_stock';
};

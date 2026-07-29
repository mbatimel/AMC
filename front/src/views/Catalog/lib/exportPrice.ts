import type { ListProductsParams, ProductListItem } from '@/core/shared/api/products';

import { fetchAllProductsRequest } from '@/core/shared/api/products';

const escapeCsv = (value: number | string): string => {
  const text = String(value);

  if (text.includes(';') || text.includes('"') || text.includes('\n')) {
    return `"${text.replaceAll('"', '""')}"`;
  }

  return text;
};

const toCsv = (products: ProductListItem[]): string => {
  const header = [
    'Артикул',
    'Наименование',
    'ГОСТ',
    'Материал',
    'Размер',
    'Упаковка',
    'Остаток',
    'Базовая цена',
    'Клиентская цена',
  ];

  const rows = products.map((product) =>
    [
      product.sku,
      product.name,
      product.gost ?? '',
      product.material ?? '',
      product.size ?? '',
      product.package_qty,
      product.stock_qty,
      product.base_price,
      product.client_price,
    ]
      .map(escapeCsv)
      .join(';'),
  );

  return `\uFEFF${[header.join(';'), ...rows].join('\n')}`;
};

export const downloadPriceList = async (params: ListProductsParams): Promise<number> => {
  const products = await fetchAllProductsRequest(params);

  if (products.length === 0) {
    throw new Error('Нет данных для выгрузки');
  }

  const blob = new Blob([toCsv(products)], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');

  link.href = url;
  link.download = 'price-list.csv';
  link.click();
  URL.revokeObjectURL(url);

  return products.length;
};

import { createEffect, createEvent, createStore, sample } from 'effector';

import type {
  Brand,
  Category,
  Product,
  ProductListItem,
  ProductWritePayload,
} from '@/core/shared/api/products';

import { $adminUserId } from '@/core/entities/adminSession';
import { writeAuditEntry } from '@/core/shared/api/admin';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import {
  createProductRequest,
  deleteProductRequest,
  getProductRequest,
  listBrandsRequest,
  listCategoriesRequest,
  listProductsRequest,
  updateProductRequest,
} from '@/core/shared/api/products';
import { toastShown } from '@/core/shared/ui/Toast/model';

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

export const adminCatalogOpened = createEvent();
export const adminProductQueryChanged = createEvent<string>();
export const adminProductOpened = createEvent<string>();
export const adminProductSaved = createEvent<{
  payload: ProductWritePayload;
  productId: null | string;
}>();
export const adminProductDeleted = createEvent<string>();

export const fetchAdminProductsFx = createEffect(async (query: string) =>
  listProductsRequest({ limit: 100, offset: 0, q: query || undefined }),
);

export const fetchAdminCategoriesFx = createEffect(() => listCategoriesRequest());
export const fetchAdminBrandsFx = createEffect(() => listBrandsRequest());

export const fetchAdminProductFx = createEffect(async (productId: string) =>
  getProductRequest(productId),
);

export const saveProductFx = createEffect(
  async ({
    payload,
    productId,
    userId,
  }: {
    payload: ProductWritePayload;
    productId: null | string;
    userId: string;
  }) => {
    if (productId) {
      await updateProductRequest(userId, productId, payload);
      await writeAuditEntry(`Обновлена карточка товара ${payload.sku}`);

      return;
    }

    await createProductRequest(userId, payload);
    await writeAuditEntry(`Создан товар ${payload.sku}`);
  },
);

export const deleteProductFx = createEffect(
  async ({ productId, userId }: { productId: string; userId: string }) => {
    await deleteProductRequest(userId, productId);
    await writeAuditEntry(`Удалён товар ${productId}`);
  },
);

export const $adminProducts = createStore<ProductListItem[]>([]).on(
  fetchAdminProductsFx.doneData,
  (_, result) => result.items,
);

export const $adminProductsTotal = createStore(0).on(
  fetchAdminProductsFx.doneData,
  (_, result) => result.pagination.total,
);

export const $adminCategories = createStore<Category[]>([]).on(
  fetchAdminCategoriesFx.doneData,
  (_, items) => items,
);

export const $adminBrands = createStore<Brand[]>([]).on(
  fetchAdminBrandsFx.doneData,
  (_, items) => items,
);

export const $adminProduct = createStore<null | Product>(null)
  .on(fetchAdminProductFx.doneData, (_, product) => product)
  .reset(fetchAdminProductFx.fail);

export const $adminProductQuery = createStore('').on(
  adminProductQueryChanged,
  (_, query) => query,
);

export const $isAdminCatalogPending = fetchAdminProductsFx.pending;
export const $isAdminProductPending = fetchAdminProductFx.pending;
export const $isProductSaving = saveProductFx.pending;

export const $adminCatalogError = createStore<null | string>(null)
  .on([fetchAdminProductsFx, saveProductFx, deleteProductFx], () => null)
  .on(
    [
      fetchAdminProductsFx.failData,
      fetchAdminProductFx.failData,
      saveProductFx.failData,
      deleteProductFx.failData,
    ],
    (_, error) => toDisplayErrorMessage(error, 'Не удалось выполнить операцию с каталогом'),
  );

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: adminCatalogOpened,
  target: [fetchAdminCategoriesFx, fetchAdminBrandsFx],
});

sample({
  clock: [adminCatalogOpened, adminProductQueryChanged],
  source: $adminProductQuery,
  target: fetchAdminProductsFx,
});

sample({
  clock: adminProductOpened,
  filter: (productId) => productId.length > 0 && productId !== 'new',
  target: fetchAdminProductFx,
});

sample({
  clock: adminProductSaved,
  source: $adminUserId,
  filter: isUserId,
  fn: (userId, { payload, productId }) => ({ payload, productId, userId }),
  target: saveProductFx,
});

sample({
  clock: adminProductDeleted,
  source: $adminUserId,
  filter: isUserId,
  fn: (userId, productId) => ({ productId, userId }),
  target: deleteProductFx,
});

sample({
  clock: [saveProductFx.done, deleteProductFx.done],
  source: $adminProductQuery,
  target: fetchAdminProductsFx,
});

sample({
  clock: saveProductFx.done,
  fn: () => ({ message: 'Товар сохранён', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: deleteProductFx.done,
  fn: () => ({ message: 'Товар удалён', tone: 'success' as const }),
  target: toastShown,
});

/* eslint-enable perfectionist/sort-objects */

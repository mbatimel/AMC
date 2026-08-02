'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

import type { ProductWritePayload } from '@/core/shared/api/products';

import { $adminUserId } from '@/core/entities/adminSession';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { deleteProductImageRequest, uploadProductImageRequest } from '@/core/shared/api/products';
import { AppPath } from '@/core/shared/router/paths';
import { FormSelect } from '@/core/shared/ui/FormSelect';

import styles from './Admin.module.css';
import {
  $adminBrands,
  $adminCatalogError,
  $adminCategories,
  $adminProduct,
  $isAdminProductPending,
  $isProductSaving,
  adminCatalogOpened,
  adminProductOpened,
  adminProductSaved,
  saveProductFx,
} from './model/catalog';
import { AdminPageHeader } from './ui/AdminPageHeader';

type AdminProductPageProps = {
  productId: string;
};

const emptyProduct = (): ProductWritePayload => ({
  basePrice: 0,
  brandID: '',
  categoryID: '',
  description: '',
  discountPercent: 0,
  gost: '',
  images: [],
  isPublished: true,
  material: '',
  name: '',
  packageQty: 1,
  size: '',
  sku: '',
  stockQty: 0,
});

export const AdminProductPage = ({ productId }: AdminProductPageProps): JSX.Element => {
  const router = useRouter();
  const isNew = productId === 'new';
  const [product, categories, brands, isPending, isSaving, error, openCatalog, openProduct, save] =
    useUnit([
      $adminProduct,
      $adminCategories,
      $adminBrands,
      $isAdminProductPending,
      $isProductSaving,
      $adminCatalogError,
      adminCatalogOpened,
      adminProductOpened,
      adminProductSaved,
    ]);
  const [draft, setDraft] = useState<ProductWritePayload>(emptyProduct);
  const adminUserId = useUnit($adminUserId);
  const [imageError, setImageError] = useState<null | string>(null);
  const [isImageSaving, setIsImageSaving] = useState(false);
  const [loadedProductId, setLoadedProductId] = useState<null | string>(null);

  useEffect(() => {
    openCatalog();
    openProduct(productId);
  }, [openCatalog, openProduct, productId]);

  if (!isNew && product && product.id === productId && loadedProductId !== productId) {
    setLoadedProductId(productId);
    setDraft({
      basePrice: product.base_price,
      brandID: product.brand_id ?? '',
      categoryID: product.category_id ?? '',
      description: product.description ?? '',
      discountPercent: product.discount_percent,
      gost: product.gost ?? '',
      images: (product.images ?? []).map((image) => ({ url: image.url })),
      isPublished: product.is_published !== false,
      material: product.material ?? '',
      name: product.name,
      packageQty: product.package_qty,
      size: product.size ?? '',
      sku: product.sku,
      stockQty: product.stock_qty,
    });
  }

  const patch = (value: Partial<ProductWritePayload>): void => {
    setDraft((previous) => ({ ...previous, ...value }));
  };

  const finalPrice = draft.basePrice * (1 - draft.discountPercent / 100);

  const submit = (): void => {
    save({
      payload: draft,
      productId: isNew ? null : productId,
    });
  };

  const uploadImage = async (file: File): Promise<void> => {
    if (!adminUserId || isNew) {
      return;
    }
    setImageError(null);
    setIsImageSaving(true);
    try {
      await uploadProductImageRequest(adminUserId, productId, file);
      openProduct(productId);
    } catch (uploadError) {
      setImageError(toDisplayErrorMessage(uploadError, 'Не удалось загрузить изображение'));
    } finally {
      setIsImageSaving(false);
    }
  };

  const removeImage = async (imageId: string): Promise<void> => {
    if (!adminUserId || isNew) {
      return;
    }
    setImageError(null);
    try {
      await deleteProductImageRequest(adminUserId, productId, imageId);
      openProduct(productId);
    } catch (deleteError) {
      setImageError(toDisplayErrorMessage(deleteError, 'Не удалось удалить изображение'));
    }
  };

  useEffect(() => {
    const unsubscribe = saveProductFx.done.watch(() => {
      router.push(AppPath.Products);
    });

    return () => {
      unsubscribe();
    };
  }, [router]);

  return (
    <>
      <AdminPageHeader
        actions={
          <>
            <Link className={clsx(styles.smallButton)} href={AppPath.Products}>
              К списку
            </Link>
            <Button isDisabled={isSaving} onPress={submit} variant="primary">
              {isSaving ? 'Сохраняем…' : 'Сохранить'}
            </Button>
          </>
        }
        subtitle={isNew ? 'Новая карточка каталога' : `ID ${productId}`}
        title={isNew ? 'Новый товар' : draft.name || 'Карточка товара'}
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}
      {isPending && !isNew ? <p className={clsx(styles.hint)}>Загружаем карточку…</p> : null}

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Основное</h2>
        <div className={clsx(styles.formGrid)}>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-sku">
              Артикул
            </label>
            <input
              className={clsx(styles.input)}
              id="product-sku"
              onChange={(event) => patch({ sku: event.target.value })}
              value={draft.sku}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-name">
              Наименование
            </label>
            <input
              className={clsx(styles.input)}
              id="product-name"
              onChange={(event) => patch({ name: event.target.value })}
              value={draft.name}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-gost">
              ГОСТ / ТУ
            </label>
            <input
              className={clsx(styles.input)}
              id="product-gost"
              onChange={(event) => patch({ gost: event.target.value })}
              value={draft.gost}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-material">
              Материал
            </label>
            <input
              className={clsx(styles.input)}
              id="product-material"
              onChange={(event) => patch({ material: event.target.value })}
              value={draft.material}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-size">
              Размер
            </label>
            <input
              className={clsx(styles.input)}
              id="product-size"
              onChange={(event) => patch({ size: event.target.value })}
              value={draft.size}
            />
          </div>
          <FormSelect
            ariaLabel="Категория товара"
            label="Категория"
            onChange={(categoryID) => patch({ categoryID })}
            options={categories.map((category) => ({
              label: category.name,
              value: category.id,
            }))}
            placeholder="Не выбрана"
            value={draft.categoryID ?? ''}
          />
          <FormSelect
            ariaLabel="Бренд товара"
            label="Бренд"
            onChange={(brandID) => patch({ brandID })}
            options={brands.map((brand) => ({ label: brand.name, value: brand.id }))}
            placeholder="Не выбран"
            value={draft.brandID ?? ''}
          />
        </div>

        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="product-description">
            Описание для портала
          </label>
          <textarea
            className={clsx(styles.textarea)}
            id="product-description"
            onChange={(event) => patch({ description: event.target.value })}
            value={draft.description}
          />
        </div>
      </section>

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Цены и остатки</h2>
        <div className={clsx(styles.formGrid)}>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-base-price">
              Базовая цена, ₽
            </label>
            <input
              className={clsx(styles.input)}
              id="product-base-price"
              min={0}
              onChange={(event) => patch({ basePrice: Number(event.target.value) || 0 })}
              step="0.01"
              type="number"
              value={draft.basePrice}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-discount">
              Скидка (акция), %
            </label>
            <input
              className={clsx(styles.input)}
              id="product-discount"
              max={100}
              min={0}
              onChange={(event) => patch({ discountPercent: Number(event.target.value) || 0 })}
              type="number"
              value={draft.discountPercent}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-client-price">
              Финальная цена, ₽
            </label>
            <input
              className={clsx(styles.input)}
              disabled
              id="product-client-price"
              readOnly
              value={finalPrice.toFixed(2)}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-package">
              Кратность упаковки
            </label>
            <input
              className={clsx(styles.input)}
              id="product-package"
              min={1}
              onChange={(event) => patch({ packageQty: Number(event.target.value) || 1 })}
              type="number"
              value={draft.packageQty}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-stock">
              Остаток
            </label>
            <input
              className={clsx(styles.input)}
              id="product-stock"
              min={0}
              onChange={(event) => patch({ stockQty: Number(event.target.value) || 0 })}
              type="number"
              value={draft.stockQty}
            />
          </div>
        </div>
      </section>

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Фотографии и публикация</h2>
        {imageError ? <p className={clsx(styles.error)}>{imageError}</p> : null}
        {!isNew ? (
          <div className={clsx(styles.listEditor)}>
            {(product?.images ?? []).map((image) => (
              <div className={clsx(styles.bannerHeader)} key={image.id}>
                <a href={image.url} rel="noreferrer" target="_blank">
                  {image.is_primary ? 'Главное фото' : 'Фото товара'}
                </a>
                <button
                  className={clsx(styles.smallButton, styles.smallButtonDanger)}
                  onClick={() => void removeImage(image.id)}
                  type="button"
                >
                  Удалить
                </button>
              </div>
            ))}
            <label className={clsx(styles.label)} htmlFor="product-image-file">
              {isImageSaving ? 'Загружаем…' : 'Добавить изображение'}
            </label>
            <input
              accept="image/jpeg,image/png,image/webp"
              disabled={isImageSaving}
              id="product-image-file"
              onChange={(event) => {
                const file = event.target.files?.[0];

                if (file) {
                  void uploadImage(file);
                }
                event.target.value = '';
              }}
              type="file"
            />
          </div>
        ) : (
          <p className={clsx(styles.hint)}>Сначала сохраните товар, затем добавьте фотографии.</p>
        )}
        <label className={clsx(styles.checkboxRow)}>
          <input
            checked={draft.isPublished}
            onChange={(event) => patch({ isPublished: event.target.checked })}
            type="checkbox"
          />
          Показывать товар в каталоге
        </label>
      </section>
    </>
  );
};

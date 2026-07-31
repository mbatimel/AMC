'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

import type { ProductWritePayload } from '@/core/shared/api/products';

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
  clientPrice: 0,
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
  const [imagesText, setImagesText] = useState('');

  useEffect(() => {
    openCatalog();
    openProduct(productId);
  }, [openCatalog, openProduct, productId]);

  useEffect(() => {
    if (!isNew && product && product.id === productId) {
      setDraft({
        basePrice: product.base_price,
        brandID: product.brand_id ?? '',
        categoryID: product.category_id ?? '',
        clientPrice: product.client_price,
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
      setImagesText((product.images ?? []).map((image) => image.url).join('\n'));
    }
  }, [isNew, product, productId]);

  const patch = (value: Partial<ProductWritePayload>): void => {
    setDraft((previous) => ({ ...previous, ...value }));
  };

  const submit = (): void => {
    const images = imagesText
      .split('\n')
      .map((url) => url.trim())
      .filter(Boolean)
      .map((url, index) => ({ is_primary: index === 0, sort_order: index, url }));

    save({
      payload: { ...draft, images },
      productId: isNew ? null : productId,
    });
  };

  useEffect(() => {
    const unsubscribe = saveProductFx.done.watch(() => {
      router.push(AppPath.AdminProducts);
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
            <Link className={clsx(styles.smallButton)} href={AppPath.AdminProducts}>
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
            <label className={clsx(styles.label)} htmlFor="product-client-price">
              Клиентская цена, ₽
            </label>
            <input
              className={clsx(styles.input)}
              id="product-client-price"
              min={0}
              onChange={(event) => patch({ clientPrice: Number(event.target.value) || 0 })}
              step="0.01"
              type="number"
              value={draft.clientPrice}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="product-discount">
              Скидка, %
            </label>
            <input
              className={clsx(styles.input)}
              id="product-discount"
              min={0}
              onChange={(event) => patch({ discountPercent: Number(event.target.value) || 0 })}
              type="number"
              value={draft.discountPercent}
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
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="product-images">
            Ссылки на изображения — по одной в строке
          </label>
          <textarea
            className={clsx(styles.textarea)}
            id="product-images"
            onChange={(event) => setImagesText(event.target.value)}
            placeholder="https://cdn.example.com/product-1.jpg"
            value={imagesText}
          />
        </div>
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

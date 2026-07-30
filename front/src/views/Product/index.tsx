'use client';

import { Breadcrumbs, Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';

import { useCart } from '@/core/entities/cart';
import { useFavorites } from '@/core/entities/favorites';
import {
  IconCart,
  IconCertificates,
  IconFavorite,
  IconSupport,
  IconTruck,
} from '@/core/shared/icons';
import { formatPrice, normalizeQtyToPackage } from '@/core/shared/lib/formatPrice';
import { getStockLevel } from '@/core/shared/lib/stock';
import { AppPath } from '@/core/shared/router/paths';
import { Page } from '@/core/shared/ui/Page';
import { ProductImageFallback } from '@/core/shared/ui/ProductImageFallback';
import { QuantityStepper } from '@/core/shared/ui/QuantityStepper';
import { toastShown } from '@/core/shared/ui/Toast/model';

import { CatalogProductCard } from '../Catalog/ui/CatalogProductCard';
import {
  $isProductPending,
  $product,
  $productError,
  $relatedProducts,
  productOpened,
} from './model';
import styles from './Product.module.css';
import { ProductSkeleton } from './ui/ProductSkeleton';

type ProductTab = 'certs' | 'delivery' | 'specs';

export const ProductPage = (): JSX.Element => {
  const params = useParams<{ id: string }>();
  const productId = params.id;
  const { addToCart } = useCart();
  const favorites = useFavorites();
  const [product, related, error, isPending, openProduct, showToast] = useUnit([
    $product,
    $relatedProducts,
    $productError,
    $isProductPending,
    productOpened,
    toastShown,
  ]);

  const [qty, setQty] = useState(1);
  const [activeImage, setActiveImage] = useState(0);
  const [imageFailed, setImageFailed] = useState(false);
  const [tab, setTab] = useState<ProductTab>('specs');
  const [syncedProductId, setSyncedProductId] = useState<null | string>(null);

  useEffect(() => {
    openProduct(productId);
  }, [openProduct, productId]);

  if (product && product.id !== syncedProductId) {
    setSyncedProductId(product.id);
    setQty(product.package_qty || 1);
    setActiveImage(0);
    setImageFailed(false);
    setTab('specs');
  }

  const images = product?.images ?? [];
  const currentImage = images[activeImage];
  const stockLevel = product ? getStockLevel(product.stock_qty) : 'out';
  const isOut = stockLevel === 'out';
  const showDiscount = Boolean(
    product &&
    (product.discount_percent > 0 ||
      (product.base_price > 0 && product.base_price > product.client_price)),
  );

  const handleAdd = (): void => {
    if (!product) {
      return;
    }

    if (qty <= 0) {
      showToast({ message: 'Укажите количество', tone: 'error' });

      return;
    }

    addToCart({
      name: product.name,
      productID: product.id,
      qty: normalizeQtyToPackage(qty, product.package_qty),
    });
  };

  const specRows = product
    ? [
        { label: 'Категория', value: product.category_name || '—' },
        { label: 'Бренд', value: product.brand_name || '—' },
        { label: 'Материал', value: product.material || '—' },
        { label: 'Размер / диаметр', value: product.size || '—' },
        { label: 'ГОСТ / ТУ', value: product.gost || '—' },
        { label: 'Упаковка', value: `${product.package_qty} шт` },
        { label: 'Артикул', value: product.sku || '—' },
      ]
    : [];

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <div className={clsx(styles.container)}>
          <Breadcrumbs className={clsx(styles.breadcrumbs)}>
            <Breadcrumbs.Item href={AppPath.Home}>Главная</Breadcrumbs.Item>
            <Breadcrumbs.Item href={AppPath.Catalog}>Каталог</Breadcrumbs.Item>
            {product?.category_name ? (
              <Breadcrumbs.Item>{product.category_name}</Breadcrumbs.Item>
            ) : null}
            {product ? <Breadcrumbs.Item>{product.name}</Breadcrumbs.Item> : null}
          </Breadcrumbs>

          {isPending ? <ProductSkeleton /> : null}
          {error ? <p className={clsx(styles.error)}>{error}</p> : null}

          {product ? (
            <>
              <section className={clsx(styles.hero)}>
                <div className={clsx(styles.gallery)}>
                  <div className={clsx(styles.mainImage)}>
                    {!currentImage || imageFailed ? (
                      <ProductImageFallback categoryName={product.category_name} />
                    ) : (
                      <img
                        alt={product.name}
                        className={clsx(styles.image)}
                        onError={() => setImageFailed(true)}
                        src={currentImage.url}
                      />
                    )}
                  </div>
                  {images.length > 1 ? (
                    <div className={clsx(styles.thumbs)}>
                      {images.map((image, index) => (
                        <button
                          className={clsx(
                            styles.thumb,
                            index === activeImage && styles.thumbActive,
                          )}
                          key={image.id}
                          onClick={() => {
                            setActiveImage(index);
                            setImageFailed(false);
                          }}
                          type="button"
                        >
                          <img alt="" src={image.url} />
                        </button>
                      ))}
                    </div>
                  ) : null}
                </div>

                <div className={clsx(styles.info)}>
                  <p className={clsx(styles.sku)}>Артикул: {product.sku}</p>
                  <h1 className={clsx(styles.title)}>{product.name}</h1>

                  <div className={clsx(styles.priceBlock)}>
                    <div className={clsx(styles.priceRow)}>
                      <span className={clsx(styles.price)}>
                        {formatPrice(product.client_price)}
                      </span>
                      {showDiscount ? (
                        <>
                          <span className={clsx(styles.oldPrice)}>
                            {formatPrice(product.base_price)}
                          </span>
                          {product.discount_percent > 0 ? (
                            <span className={clsx(styles.discount)}>
                              -{Math.round(product.discount_percent)}%
                            </span>
                          ) : null}
                        </>
                      ) : null}
                    </div>
                    <span className={clsx(styles.priceHint)}>
                      {showDiscount ? 'Ваша спеццена' : 'Ваша оптовая цена'}
                    </span>
                  </div>

                  <div
                    className={clsx(
                      styles.stockRow,
                      stockLevel === 'out' && styles.stockOut,
                      stockLevel === 'low' && styles.stockLow,
                    )}
                  >
                    <span>{isOut ? 'Нет в наличии' : `В наличии: ${product.stock_qty}`}</span>
                  </div>

                  <div className={clsx(styles.buyRow)}>
                    <QuantityStepper
                      className={clsx(styles.stepper)}
                      disabled={isOut}
                      onChange={setQty}
                      step={product.package_qty}
                      value={qty}
                    />
                    <Button
                      className={clsx(styles.cartButton)}
                      isDisabled={isOut}
                      onPress={handleAdd}
                      variant="primary"
                    >
                      {isOut ? (
                        'Нет'
                      ) : (
                        <>
                          <IconCart currentColor="currentColor" height={16} width={16} />В корзину
                        </>
                      )}
                    </Button>
                    <Button
                      aria-label="Избранное"
                      className={clsx(
                        styles.favorite,
                        favorites.isFavorite(product.id) && styles.favoriteActive,
                      )}
                      isIconOnly
                      onPress={() => favorites.toggleFavorite(product.id)}
                      variant="outline"
                    >
                      <IconFavorite currentColor="currentColor" height={18} width={18} />
                    </Button>
                  </div>

                  <div className={clsx(styles.infoCards)}>
                    <div className={clsx(styles.infoCard)}>
                      <span aria-hidden className={clsx(styles.infoIcon)}>
                        <IconTruck currentColor="currentColor" height={18} width={18} />
                      </span>
                      <div className={clsx(styles.infoCardBody)}>
                        <strong>Доставка</strong>
                        <span>Самовывоз и доставка ТК / собственной службой</span>
                      </div>
                    </div>
                    <div className={clsx(styles.infoCard)}>
                      <span aria-hidden className={clsx(styles.infoIcon)}>
                        <IconCertificates currentColor="currentColor" height={18} width={18} />
                      </span>
                      <div className={clsx(styles.infoCardBody)}>
                        <strong>Сертификаты</strong>
                        <span>Документы по запросу в личном кабинете</span>
                      </div>
                    </div>
                  </div>
                </div>
              </section>

              <section className={clsx(styles.tabs)}>
                <div
                  aria-label="Информация о товаре"
                  className={clsx(styles.tabList)}
                  role="tablist"
                >
                  {(
                    [
                      { id: 'specs', label: 'Характеристики' },
                      { id: 'delivery', label: 'Доставка и оплата' },
                      { id: 'certs', label: 'Сертификаты' },
                    ] as const
                  ).map((item) => (
                    <button
                      aria-selected={tab === item.id}
                      className={clsx(styles.tab, tab === item.id && styles.tabActive)}
                      key={item.id}
                      onClick={() => setTab(item.id)}
                      role="tab"
                      type="button"
                    >
                      {item.label}
                    </button>
                  ))}
                </div>

                {tab === 'specs' ? (
                  <div className={clsx(styles.tabPanel)} role="tabpanel">
                    <dl className={clsx(styles.specs)}>
                      {specRows.map((row) => (
                        <div className={clsx(styles.specRow)} key={row.label}>
                          <dt>{row.label}</dt>
                          <dd
                            className={clsx(
                              row.label === 'Бренд' && product.brand_name && styles.specBrand,
                            )}
                          >
                            {row.value}
                          </dd>
                        </div>
                      ))}
                      {product.description ? (
                        <div className={clsx(styles.specRow, styles.description)}>
                          <dt>Описание</dt>
                          <dd>{product.description}</dd>
                        </div>
                      ) : null}
                    </dl>
                  </div>
                ) : null}

                {tab === 'delivery' ? (
                  <div className={clsx(styles.tabPanel)} role="tabpanel">
                    <div className={clsx(styles.staticGrid)}>
                      <div>
                        <h3>Самовывоз</h3>
                        <p>Со склада в согласованный день после подтверждения заказа.</p>
                      </div>
                      <div>
                        <h3>ТК</h3>
                        <p>Отправка транспортными компаниями по тарифам перевозчика.</p>
                      </div>
                      <div>
                        <h3>Собственная доставка</h3>
                        <p>Доставка силами компании при выполнении условий по сумме заказа.</p>
                      </div>
                    </div>
                  </div>
                ) : null}

                {tab === 'certs' ? (
                  <div className={clsx(styles.tabPanel)} role="tabpanel">
                    <div className={clsx(styles.staticGrid)}>
                      <div>
                        <h3>{product.gost || 'ГОСТ / ТУ'}</h3>
                        <p>Документы предоставляются по запросу.</p>
                      </div>
                      <div>
                        <h3>ISO 9001</h3>
                        <p>Сертификат системы менеджмента качества — по запросу.</p>
                      </div>
                    </div>
                  </div>
                ) : null}
              </section>

              {related.length > 0 ? (
                <section className={clsx(styles.related)}>
                  <h2 className={clsx(styles.relatedTitle)}>Сопутствующие товары</h2>
                  <div className={clsx(styles.relatedGrid)}>
                    {related.map((item) => (
                      <CatalogProductCard
                        isFavorite={favorites.isFavorite(item.id)}
                        key={item.id}
                        onAddToCart={() =>
                          addToCart({
                            name: item.name,
                            productID: item.id,
                            qty: item.package_qty,
                          })
                        }
                        onToggleFavorite={() => favorites.toggleFavorite(item.id)}
                        product={item}
                      />
                    ))}
                  </div>
                </section>
              ) : null}

              <section className={clsx(styles.consult)}>
                <div className={clsx(styles.consultText)}>
                  <h2>Нужна консультация по подбору?</h2>
                  <p>Поможем подобрать инструмент под задачу и условия работы.</p>
                </div>
                <div className={clsx(styles.consultActions)}>
                  <Link className={clsx(styles.consultPrimary)} href={AppPath.Catalog}>
                    Подбор инструмента
                  </Link>
                  <a className={clsx(styles.consultSecondary)} href="#support">
                    <IconSupport currentColor="currentColor" height={16} width={16} />
                    Поддержка
                  </a>
                </div>
              </section>
            </>
          ) : null}
        </div>
      </div>
    </Page>
  );
};

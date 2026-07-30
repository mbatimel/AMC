'use client';

import { Breadcrumbs, Button, Chip, Tab, TabList, TabPanel, Tabs } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';

import { useCart } from '@/core/entities/cart';
import { useFavorites } from '@/core/entities/favorites';
import { formatPrice, normalizeQtyToPackage } from '@/core/shared/lib/formatPrice';
import { getStockLevel } from '@/core/shared/lib/stock';
import { AppPath } from '@/core/shared/router/paths';
import { Page } from '@/core/shared/ui/Page';
import { ProductImageFallback } from '@/core/shared/ui/ProductImageFallback';
import { QuantityStepper } from '@/core/shared/ui/QuantityStepper';
import { toastShown } from '@/core/shared/ui/Toast/model';

import { $isProductPending, $product, $productError, productOpened } from './model';
import styles from './Product.module.css';

type ProductTab = 'certs' | 'delivery' | 'specs';

export const ProductPage = (): JSX.Element => {
  const params = useParams<{ id: string }>();
  const productId = params.id;
  const { addToCart } = useCart();
  const favorites = useFavorites();
  const [product, error, isPending, openProduct, showToast] = useUnit([
    $product,
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
  const hasDiscount = Boolean(product && product.discount_percent > 0);

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

          {isPending ? <p className={clsx(styles.status)}>Загрузка товара…</p> : null}
          {error ? <p className={clsx(styles.error)}>{error}</p> : null}

          {product ? (
            <>
              <div className={clsx(styles.hero)}>
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
                        <Button
                          className={clsx(
                            styles.thumb,
                            index === activeImage && styles.thumbActive,
                          )}
                          isIconOnly
                          key={image.id}
                          onPress={() => {
                            setActiveImage(index);
                            setImageFailed(false);
                          }}
                          variant="outline"
                        >
                          <img alt="" src={image.url} />
                        </Button>
                      ))}
                    </div>
                  ) : null}
                </div>

                <div className={clsx(styles.info)}>
                  <p className={clsx(styles.sku)}>Арт.: {product.sku}</p>
                  <h1 className={clsx(styles.title)}>{product.name}</h1>

                  <div className={clsx(styles.priceBlock)}>
                    <span className={clsx(styles.price)}>{formatPrice(product.client_price)}</span>
                    {hasDiscount ? (
                      <>
                        <span className={clsx(styles.oldPrice)}>
                          {formatPrice(product.base_price)}
                        </span>
                        <Chip color="warning" size="sm">
                          <Chip.Label>-{Math.round(product.discount_percent)}%</Chip.Label>
                        </Chip>
                      </>
                    ) : null}
                    <span className={clsx(styles.priceHint)}>Ваша оптовая цена</span>
                  </div>

                  <Chip
                    color={isOut ? 'danger' : stockLevel === 'low' ? 'warning' : 'success'}
                    size="sm"
                  >
                    <Chip.Label>
                      {isOut ? 'Нет в наличии' : `В наличии: ${product.stock_qty}`}
                    </Chip.Label>
                  </Chip>

                  <div className={clsx(styles.buyRow)}>
                    <QuantityStepper
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
                      {isOut ? 'Нет' : 'В корзину'}
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
                      ★
                    </Button>
                  </div>

                  <div className={clsx(styles.infoCards)}>
                    <div className={clsx(styles.infoCard)}>
                      <strong>Доставка</strong>
                      <span>Самовывоз и доставка ТК / собственной службой</span>
                    </div>
                    <div className={clsx(styles.infoCard)}>
                      <strong>Сертификаты</strong>
                      <span>Документы по запросу в личном кабинете</span>
                    </div>
                  </div>
                </div>
              </div>

              <Tabs
                className={clsx(styles.tabs)}
                onSelectionChange={(key) => setTab(String(key) as ProductTab)}
                selectedKey={tab}
              >
                <TabList aria-label="Информация о товаре" className={clsx(styles.tabList)}>
                  <Tab className={clsx(styles.tab)} id="specs">
                    Характеристики
                  </Tab>
                  <Tab className={clsx(styles.tab)} id="delivery">
                    Доставка и оплата
                  </Tab>
                  <Tab className={clsx(styles.tab)} id="certs">
                    Сертификаты
                  </Tab>
                </TabList>

                <TabPanel className={clsx(styles.tabPanel)} id="specs">
                  <dl className={clsx(styles.specs)}>
                    <div>
                      <dt>Категория</dt>
                      <dd>{product.category_name || '—'}</dd>
                    </div>
                    <div>
                      <dt>Бренд</dt>
                      <dd>{product.brand_name || '—'}</dd>
                    </div>
                    <div>
                      <dt>Материал</dt>
                      <dd>{product.material || '—'}</dd>
                    </div>
                    <div>
                      <dt>Размер</dt>
                      <dd>{product.size || '—'}</dd>
                    </div>
                    <div>
                      <dt>ГОСТ</dt>
                      <dd>{product.gost || '—'}</dd>
                    </div>
                    <div>
                      <dt>Упаковка</dt>
                      <dd>{product.package_qty} шт</dd>
                    </div>
                    {product.description ? (
                      <div className={clsx(styles.description)}>
                        <dt>Описание</dt>
                        <dd>{product.description}</dd>
                      </div>
                    ) : null}
                  </dl>
                </TabPanel>

                <TabPanel className={clsx(styles.tabPanel)} id="delivery">
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
                </TabPanel>

                <TabPanel className={clsx(styles.tabPanel)} id="certs">
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
                </TabPanel>
              </Tabs>
            </>
          ) : null}
        </div>
      </div>
    </Page>
  );
};

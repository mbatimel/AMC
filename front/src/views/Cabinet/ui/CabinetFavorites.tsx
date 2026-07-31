'use client';

import { Alert, Button, Description, EmptyState, Spinner, Typography } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

import { useCart } from '@/core/entities/cart';
import { useFavorites } from '@/core/entities/favorites';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { AppPath, getProductPath } from '@/core/shared/router/paths';

import styles from '../Cabinet.module.css';
import extraStyles from '../CabinetExtras.module.css';
import {
  $favoritesError,
  $isFavoritesPending,
  $visibleFavorites,
  cabinetFavoritesOpened,
} from '../model/favorites';

export const CabinetFavorites = (): JSX.Element => {
  const router = useRouter();
  const { addToCart } = useCart();
  const { toggleFavorite } = useFavorites();
  const [products, isPending, error, open] = useUnit([
    $visibleFavorites,
    $isFavoritesPending,
    $favoritesError,
    cabinetFavoritesOpened,
  ]);

  useEffect(() => {
    open();
  }, [open]);

  return (
    <div className={clsx(styles.main)}>
      <div className={clsx(styles.pageHeader)}>
        <div>
          <Typography.Heading className={clsx(styles.pageTitle)} level={1}>
            Избранное
          </Typography.Heading>
          <Description className={clsx(styles.pageSubtitle)}>
            {products.length > 0
              ? `${products.length} позиций в списке`
              : 'Сохранённые позиции каталога'}
          </Description>
        </div>
      </div>

      {error ? (
        <Alert className={clsx(styles.alert)} status="danger">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      ) : null}

      {isPending && products.length === 0 ? <Spinner /> : null}

      {!isPending && products.length === 0 ? (
        <EmptyState className={clsx(extraStyles.emptyState)}>
          <Description>
            Пока пусто. Добавляйте позиции в избранное из каталога — они появятся здесь.
          </Description>
          <Button onPress={() => router.push(AppPath.Catalog)} variant="primary">
            Перейти в каталог
          </Button>
        </EmptyState>
      ) : null}

      {products.length > 0 ? (
        <ul className={clsx(extraStyles.cards)}>
          {products.map((product) => (
            <li className={clsx(extraStyles.card)} key={product.id}>
              <div className={clsx(extraStyles.cardMain)}>
                <p className={clsx(extraStyles.cardSku)}>
                  {product.sku}
                  {product.gost ? ` · ${product.gost}` : ''}
                </p>
                <p className={clsx(extraStyles.cardName)}>{product.name}</p>
                <p className={clsx(extraStyles.cardMeta)}>
                  {product.stock_qty > 0 ? `В наличии: ${product.stock_qty}` : 'Нет в наличии'}
                  {product.brand_name ? ` · ${product.brand_name}` : ''}
                </p>
              </div>

              <div className={clsx(extraStyles.cardSide)}>
                <span className={clsx(extraStyles.cardPrice)}>
                  {formatPrice(product.client_price || product.base_price)}
                </span>
                <div className={clsx(extraStyles.cardActions)}>
                  <Button
                    onPress={() => router.push(getProductPath(product.id))}
                    variant="outline"
                  >
                    Открыть
                  </Button>
                  <Button
                    isDisabled={product.stock_qty <= 0}
                    onPress={() =>
                      addToCart({
                        name: product.name,
                        productID: product.id,
                        qty: product.package_qty,
                      })
                    }
                    variant="primary"
                  >
                    В корзину
                  </Button>
                  <Button onPress={() => toggleFavorite(product.id)} variant="outline">
                    Убрать
                  </Button>
                </div>
              </div>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
};

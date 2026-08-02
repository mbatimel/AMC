'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import { IconCalendar } from '@/core/shared/icons/IconCalendar';
import { IconLayers } from '@/core/shared/icons/IconLayers';
import { IconPackage } from '@/core/shared/icons/IconPackage';
import { IconPriceTag } from '@/core/shared/icons/IconPriceTag';
import { getCatalogPromotionPath } from '@/core/shared/router/paths';
import { InfoCard, InfoPage, InfoPageSkeleton } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import {
  formatProductsCount,
  formatPromoEndsAt,
  promoConditionText,
  promoNeedsQty,
} from './lib/format';
import { $activePromos, $isPromoPending, $promoError, promoMounted } from './model';
import styles from './Promo.module.css';

export const Promo = (): JSX.Element => {
  const [promos, isPending, error, mount] = useUnit([
    $activePromos,
    $isPromoPending,
    $promoError,
    promoMounted,
  ]);

  useEffect(() => {
    mount();
  }, [mount]);

  const hasPromos = promos.length > 0;

  return (
    <Page>
      <InfoPage
        description="Сниженные цены и скидки от количества. Нажмите на акцию, чтобы увидеть участвующие товары в каталоге."
        eyebrow="Действующие акции"
        title="Акции и спецпредложения"
      >
        <InfoCard>
          {isPending && !hasPromos ? <InfoPageSkeleton /> : null}
          {error && !hasPromos ? <p className={clsx(styles.error)}>{error}</p> : null}

          {!isPending && !error && !hasPromos ? (
            <div className={clsx(styles.empty)}>
              <p className={clsx(styles.emptyTitle)}>Сейчас нет действующих акций</p>
              <p className={clsx(styles.emptyText)}>
                Загляните позже — здесь появятся актуальные предложения
              </p>
            </div>
          ) : null}

          {hasPromos ? (
            <ul className={clsx(styles.grid)}>
              {promos.map((promo) => (
                <li key={promo.id}>
                  <Link
                    className={clsx(styles.card)}
                    href={getCatalogPromotionPath(promo.id, promo.name)}
                  >
                    <span className={clsx(styles.badge)}>−{promo.discount_percent}%</span>
                    <h2 className={clsx(styles.name)}>{promo.name}</h2>
                    <div className={clsx(styles.meta)}>
                      <span className={clsx(styles.metaRow)}>
                        <IconCalendar className={clsx(styles.metaIcon)} height={13} width={13} />
                        до {formatPromoEndsAt(promo.ends_at)}
                      </span>
                      <span className={clsx(styles.metaRow)}>
                        {promoNeedsQty(promo) ? (
                          <IconLayers className={clsx(styles.metaIcon)} height={13} width={13} />
                        ) : (
                          <IconPriceTag className={clsx(styles.metaIcon)} height={13} width={13} />
                        )}
                        {promoConditionText(promo)}
                      </span>
                      <span className={clsx(styles.metaRow)}>
                        <IconPackage className={clsx(styles.metaIcon)} height={13} width={13} />
                        {formatProductsCount(promo.products.length)}
                      </span>
                    </div>
                    <span className={clsx(styles.cta)}>Смотреть товары →</span>
                  </Link>
                </li>
              ))}
            </ul>
          ) : null}
        </InfoCard>

        <InfoCard>
          <div className={clsx(styles.info)}>
            <h2 className={clsx(styles.infoTitle)}>Как работают акции</h2>
            <p className={clsx(styles.infoText)}>
              Скидка применяется в корзине при достижении минимального количества по товару. Если у
              позиции есть и ручная скидка, и акция — выбирается более выгодный вариант.
            </p>
          </div>
        </InfoCard>
      </InfoPage>
    </Page>
  );
};

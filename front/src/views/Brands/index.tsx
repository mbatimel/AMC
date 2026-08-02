'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import { AppPath } from '@/core/shared/router/paths';
import { FOOTER_BRANDS } from '@/core/shared/ui/Footer/constants';
import { InfoCard, InfoPage, InfoPageSkeleton } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import styles from './Brands.module.css';
import { $brands, $brandsError, $isBrandsPending, brandsMounted } from './model';

export const Brands = (): JSX.Element => {
  const [brands, isPending, error, mount] = useUnit([
    $brands,
    $isBrandsPending,
    $brandsError,
    brandsMounted,
  ]);

  useEffect(() => {
    mount();
  }, [mount]);

  const hasApiBrands = brands.length > 0;

  return (
    <Page>
      <InfoPage
        description="Собственное производство и партнёрские марки, представленные на портале."
        eyebrow="Каталог"
        title="Бренды"
      >
        <InfoCard>
          {isPending && !hasApiBrands ? <InfoPageSkeleton /> : null}
          {error && !hasApiBrands ? <p className={clsx(styles.error)}>{error}</p> : null}

          {hasApiBrands ? (
            <ul className={clsx(styles.grid)}>
              {brands.map((brand) => (
                <li key={brand.id}>
                  <Link className={clsx(styles.card)} href={AppPath.Catalog}>
                    <span aria-hidden className={clsx(styles.logo)}>
                      {brand.name.slice(0, 2).toUpperCase()}
                    </span>
                    <span className={clsx(styles.name)}>{brand.name}</span>
                    <span className={clsx(styles.action)}>Смотреть товары →</span>
                  </Link>
                </li>
              ))}
            </ul>
          ) : null}

          {!isPending && !hasApiBrands ? (
            <ul className={clsx(styles.grid)}>
              {FOOTER_BRANDS.map((brandName) => (
                <li key={brandName}>
                  <Link className={clsx(styles.card)} href={AppPath.Catalog}>
                    <span aria-hidden className={clsx(styles.logo)}>
                      {brandName.slice(0, 2).toUpperCase()}
                    </span>
                    <span className={clsx(styles.name)}>{brandName}</span>
                    <span className={clsx(styles.action)}>Смотреть товары →</span>
                  </Link>
                </li>
              ))}
            </ul>
          ) : null}
        </InfoCard>
      </InfoPage>
    </Page>
  );
};

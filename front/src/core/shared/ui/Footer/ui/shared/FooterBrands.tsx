'use client';

import clsx from 'clsx';
import Link from 'next/link';
import { useEffect, useState } from 'react';

import type { Brand } from '@/core/shared/api/products';

import { listBrandsRequest } from '@/core/shared/api/products';
import { IconBrands } from '@/core/shared/icons';
import { getCatalogBrandPath } from '@/core/shared/router/paths';

import { FOOTER_BRANDS_TITLE } from '../../constants';
import styles from '../../Footer.module.css';

export const FooterBrands = (): JSX.Element | null => {
  const [brands, setBrands] = useState<Brand[]>([]);

  useEffect(() => {
    let cancelled = false;

    void listBrandsRequest()
      .then((items) => {
        if (!cancelled) {
          setBrands(items);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setBrands([]);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  if (brands.length === 0) {
    return null;
  }

  return (
    <section className={clsx(styles.brandsSection)}>
      <div className={clsx(styles.container)}>
        <div className={clsx(styles.brandsTitle)}>
          <IconBrands className={clsx(styles.brandsTitleIcon)} height={13} width={13} />
          <span>{FOOTER_BRANDS_TITLE}</span>
        </div>

        <ul className={clsx(styles.brandsList)}>
          {brands.map((brand) => (
            <li className={clsx(styles.brandItem)} key={brand.id}>
              <Link
                className={clsx(styles.brandChip)}
                href={getCatalogBrandPath(brand.id, brand.name)}
              >
                <span>{brand.name}</span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
};

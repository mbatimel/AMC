import clsx from 'clsx';
import Link from 'next/link';

import { IconBrands } from '@/core/shared/icons';

import { FOOTER_BRANDS, FOOTER_BRANDS_TITLE } from '../../constants';
import styles from '../../Footer.module.css';

export const FooterBrands = (): JSX.Element => {
  return (
    <section className={clsx(styles.brandsSection)}>
      <div className={clsx(styles.container)}>
        <div className={clsx(styles.brandsTitle)}>
          <IconBrands className={clsx(styles.brandsTitleIcon)} height={13} width={13} />
          <span>{FOOTER_BRANDS_TITLE}</span>
        </div>

        <ul className={clsx(styles.brandsList)}>
          {FOOTER_BRANDS.map((brand) => (
            <li className={clsx(styles.brandItem)} key={brand}>
              <Link className={clsx(styles.brandChip)} href="#">
                <span>{brand}</span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
};

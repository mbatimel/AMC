import clsx from 'clsx';

import {
  FOOTER_CABINET_LINKS,
  FOOTER_CATALOG_LINKS,
  FOOTER_INFO_LINKS,
} from '../../constants';
import styles from '../../Footer.module.css';
import { FooterAbout } from '../shared/FooterAbout';
import { FooterBrands } from '../shared/FooterBrands';
import { FooterLegal } from '../shared/FooterLegal';
import { FooterNavColumn } from '../shared/FooterNavColumn';
import mobileStyles from './FooterMobile.module.css';

export const FooterMobile = (): JSX.Element => {
  return (
    <>
      <FooterBrands />

      <section className={clsx(mobileStyles.mainSection)}>
        <div className={clsx(styles.container, mobileStyles.mainStack)}>
          <FooterAbout />

          <div className={clsx(mobileStyles.navGroup)}>
            <FooterNavColumn items={FOOTER_CATALOG_LINKS} title="Каталог" />
            <FooterNavColumn items={FOOTER_CABINET_LINKS} title="Кабинет" />
            <FooterNavColumn items={FOOTER_INFO_LINKS} title="Информация" />
          </div>
        </div>
      </section>

      <FooterLegal />
    </>
  );
};

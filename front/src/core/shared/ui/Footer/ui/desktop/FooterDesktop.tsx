import clsx from 'clsx';

import type { FooterLinkItem } from '../../constants';

import { FOOTER_CATALOG_LINKS, FOOTER_INFO_LINKS } from '../../constants';
import styles from '../../Footer.module.css';
import { FooterAbout } from '../shared/FooterAbout';
import { FooterBrands } from '../shared/FooterBrands';
import { FooterLegal } from '../shared/FooterLegal';
import { FooterNavColumn } from '../shared/FooterNavColumn';
import desktopStyles from './FooterDesktop.module.css';

type FooterDesktopProps = {
  cabinetLinks: FooterLinkItem[];
  cabinetTitle: string;
};

export const FooterDesktop = ({ cabinetLinks, cabinetTitle }: FooterDesktopProps): JSX.Element => {
  return (
    <>
      <FooterBrands />

      <section className={clsx(desktopStyles.mainSection)}>
        <div className={clsx(styles.container, desktopStyles.mainGrid)}>
          <FooterAbout />
          <FooterNavColumn items={FOOTER_CATALOG_LINKS} title="Каталог" />
          <FooterNavColumn items={cabinetLinks} title={cabinetTitle} />
          <FooterNavColumn items={FOOTER_INFO_LINKS} title="Информация" />
        </div>
      </section>

      <FooterLegal />
    </>
  );
};

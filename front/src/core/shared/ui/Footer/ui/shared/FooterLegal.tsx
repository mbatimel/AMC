import clsx from 'clsx';
import Link from 'next/link';

import { FOOTER_COPYRIGHT, FOOTER_LEGAL_LINKS, FOOTER_VERSION } from '../../constants';
import styles from '../../Footer.module.css';

export const FooterLegal = (): JSX.Element => {
  return (
    <section className={clsx(styles.legalSection)}>
      <div className={clsx(styles.container)}>
        <ul className={clsx(styles.legalLinks)}>
          {FOOTER_LEGAL_LINKS.map((item) => (
            <li className={clsx(styles.legalItem)} key={item.label}>
              <Link className={clsx(styles.legalLink)} href={item.href}>
                <span>{item.label}</span>
              </Link>
            </li>
          ))}
        </ul>

        <div className={clsx(styles.copyrightRow)}>
          <span className={clsx(styles.copyrightText)}>{FOOTER_COPYRIGHT}</span>
          <span className={clsx(styles.versionText)}>{FOOTER_VERSION}</span>
        </div>
      </div>
    </section>
  );
};

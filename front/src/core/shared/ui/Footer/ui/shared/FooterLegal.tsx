import clsx from 'clsx';
import Link from 'next/link';

import { useContent } from '@/core/entities/content';
import { getLegalDocPath } from '@/core/shared/router/paths';

import { FOOTER_COPYRIGHT, FOOTER_LEGAL_LINKS, FOOTER_VERSION } from '../../constants';
import styles from '../../Footer.module.css';

export const FooterLegal = (): JSX.Element => {
  const { legalDocs } = useContent();
  const links =
    legalDocs.length > 0
      ? legalDocs.map((doc) => ({ href: getLegalDocPath(doc.id), label: doc.name }))
      : FOOTER_LEGAL_LINKS;

  return (
    <section className={clsx(styles.legalSection)}>
      <div className={clsx(styles.container)}>
        <ul className={clsx(styles.legalLinks)}>
          {links.map((item) => (
            <li className={clsx(styles.legalItem)} key={item.href}>
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

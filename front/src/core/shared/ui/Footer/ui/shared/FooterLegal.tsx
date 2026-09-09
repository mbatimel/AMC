import clsx from 'clsx';
import Link from 'next/link';

import { useContent } from '@/core/entities/content';
import { getLegalDocPath } from '@/core/shared/router/paths';

import { FOOTER_COPYRIGHT, FOOTER_VERSION } from '../../constants';
import styles from '../../Footer.module.css';

export const FooterLegal = (): JSX.Element => {
  const { legalDocs } = useContent();
  const links = legalDocs.map((doc) => ({ href: getLegalDocPath(doc.id), label: doc.name }));

  return (
    <section className={clsx(styles.legalSection)}>
      <div className={clsx(styles.container)}>
        {links.length > 0 ? (
          <ul className={clsx(styles.legalLinks)}>
            {links.map((item) => (
              <li className={clsx(styles.legalItem)} key={item.href}>
                <Link className={clsx(styles.legalLink)} href={item.href}>
                  <span>{item.label}</span>
                </Link>
              </li>
            ))}
          </ul>
        ) : null}

        <div className={clsx(styles.copyrightRow)}>
          <span className={clsx(styles.copyrightText)}>{FOOTER_COPYRIGHT}</span>
          <span className={clsx(styles.versionText)}>{FOOTER_VERSION}</span>
        </div>
      </div>
    </section>
  );
};

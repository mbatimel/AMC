import clsx from 'clsx';
import Link from 'next/link';

import { AppPath } from '@/core/shared/router/paths';
import { Page } from '@/core/shared/ui/Page';

import styles from './NotFound.module.css';

export const NotFound = (): JSX.Element => {
  return (
    <Page hasAssistant={false}>
      <div className={clsx(styles.root)}>
        <div className={clsx(styles.card)}>
          <p aria-hidden="true" className={clsx(styles.code)}>
            404
          </p>
          <h1 className={clsx(styles.title)}>Страница не найдена</h1>
          <p className={clsx(styles.text)}>
            Такой страницы нет или она была перемещена. Можно вернуться на главную или открыть
            каталог.
          </p>
          <div className={clsx(styles.actions)}>
            <Link className={clsx(styles.primaryLink)} href={AppPath.Home}>
              На главную
            </Link>
            <Link className={clsx(styles.secondaryLink)} href={AppPath.Catalog}>
              В каталог
            </Link>
          </div>
        </div>
      </div>
    </Page>
  );
};

import { Page } from '@/core/shared/ui/Page';

import styles from './Cabinet.module.css';

export const Cabinet = (): JSX.Element => {
  return (
    <Page>
      <div className={styles.root}>
        <h1 className={styles.title}>Личный кабинет</h1>
        <p className={styles.description}>Раздел в разработке.</p>
      </div>
    </Page>
  );
};

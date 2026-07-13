import clsx from 'clsx';

import { Footer } from '@/core/shared/ui/Footer';
import { Header } from '@/core/shared/ui/Header';

import styles from './Page.module.css';

type PageProps = {
  children: React.ReactNode;
};

export const Page = ({ children }: PageProps): JSX.Element => {
  return (
    <div className={clsx(styles.root)}>
      <Header />
      <main className={clsx(styles.main)}>{children}</main>
      <Footer />
    </div>
  );
};

import clsx from 'clsx';

import { AssistantWidget } from '@/core/shared/ui/AssistantWidget';
import { BackToTop } from '@/core/shared/ui/BackToTop';
import { Footer } from '@/core/shared/ui/Footer';
import { Header } from '@/core/shared/ui/Header';

import styles from './Page.module.css';

type PageProps = {
  children: React.ReactNode;
  /** Скрыть плавающего помощника (например, на страницах оформления заказа). */
  hasAssistant?: boolean;
};

export const Page = ({ children, hasAssistant = true }: PageProps): JSX.Element => {
  return (
    <div className={clsx(styles.root)}>
      <Header />
      <main className={clsx(styles.main)}>{children}</main>
      <Footer />
      {hasAssistant ? <AssistantWidget /> : null}
      <BackToTop />
    </div>
  );
};

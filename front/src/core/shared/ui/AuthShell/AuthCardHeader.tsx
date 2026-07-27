import clsx from 'clsx';

import type { Icon } from '@/core/shared/icons/types';

import styles from './AuthCardHeader.module.css';

type AuthCardHeaderProps = {
  description: string;
  icon: Icon;
  title: string;
};

export const AuthCardHeader = ({
  description,
  icon: IconComponent,
  title,
}: AuthCardHeaderProps): JSX.Element => {
  return (
    <header className={clsx(styles.root)}>
      <div aria-hidden className={clsx(styles.iconWrap)}>
        <IconComponent currentColor="var(--header-brand)" height={18} width={18} />
      </div>
      <h1 className={clsx(styles.title)}>{title}</h1>
      <p className={clsx(styles.description)}>{description}</p>
    </header>
  );
};

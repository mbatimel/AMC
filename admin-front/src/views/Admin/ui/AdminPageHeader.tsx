import clsx from 'clsx';

import styles from './AdminPageHeader.module.css';

type AdminPageHeaderProps = {
  actions?: React.ReactNode;
  subtitle?: string;
  title: string;
};

export const AdminPageHeader = ({
  actions,
  subtitle,
  title,
}: AdminPageHeaderProps): JSX.Element => {
  return (
    <div className={clsx(styles.pageHeader)}>
      <div>
        <h1 className={clsx(styles.pageTitle)}>{title}</h1>
        {subtitle ? <p className={clsx(styles.pageSubtitle)}>{subtitle}</p> : null}
      </div>
      {actions ? <div className={clsx(styles.pageActions)}>{actions}</div> : null}
    </div>
  );
};

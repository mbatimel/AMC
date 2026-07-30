import clsx from 'clsx';

import styles from './ProductSkeleton.module.css';

export const ProductSkeleton = (): JSX.Element => {
  return (
    <div aria-busy aria-label="Загрузка товара" className={clsx(styles.root)} role="status">
      <div className={clsx(styles.hero)}>
        <div className={clsx(styles.gallery)}>
          <div className={clsx(styles.bone, styles.mainImage)} />
          <div className={clsx(styles.thumbs)}>
            <div className={clsx(styles.bone, styles.thumb)} />
            <div className={clsx(styles.bone, styles.thumb)} />
            <div className={clsx(styles.bone, styles.thumb)} />
          </div>
        </div>

        <div className={clsx(styles.info)}>
          <div className={clsx(styles.bone, styles.sku)} />
          <div className={clsx(styles.bone, styles.title)} />
          <div className={clsx(styles.bone, styles.titleSm)} />
          <div className={clsx(styles.bone, styles.price)} />
          <div className={clsx(styles.bone, styles.hint)} />
          <div className={clsx(styles.bone, styles.stock)} />
          <div className={clsx(styles.buyRow)}>
            <div className={clsx(styles.bone, styles.stepper)} />
            <div className={clsx(styles.bone, styles.cartBtn)} />
            <div className={clsx(styles.bone, styles.favBtn)} />
          </div>
          <div className={clsx(styles.infoCards)}>
            <div className={clsx(styles.bone, styles.infoCard)} />
            <div className={clsx(styles.bone, styles.infoCard)} />
          </div>
        </div>
      </div>

      <div className={clsx(styles.tabs)}>
        <div className={clsx(styles.tabList)}>
          <div className={clsx(styles.bone, styles.tab)} />
          <div className={clsx(styles.bone, styles.tab)} />
          <div className={clsx(styles.bone, styles.tab)} />
        </div>
        <div className={clsx(styles.tabPanel)}>
          {Array.from({ length: 6 }, (_, index) => (
            <div className={clsx(styles.specRow)} key={index}>
              <div className={clsx(styles.bone, styles.specLabel)} />
              <div className={clsx(styles.bone, styles.specValue)} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

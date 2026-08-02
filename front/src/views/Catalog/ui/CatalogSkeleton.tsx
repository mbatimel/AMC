import clsx from 'clsx';

import type { CatalogViewMode } from '../lib/filters';

import styles from './CatalogSkeleton.module.css';

type CatalogSkeletonProps = {
  view: CatalogViewMode;
};

const TABLE_ROWS = 8;
const CARD_COUNT = 6;

export const CatalogSkeleton = ({ view }: CatalogSkeletonProps): JSX.Element => {
  if (view === 'cards') {
    return (
      <div aria-busy aria-label="Загрузка каталога" className={clsx(styles.cards)} role="status">
        {Array.from({ length: CARD_COUNT }, (_, index) => (
          <div className={clsx(styles.card)} key={index}>
            <div className={clsx(styles.cardMedia, styles.bone)} />
            <div className={clsx(styles.cardBody)}>
              <div className={clsx(styles.bone, styles.lineSm)} />
              <div className={clsx(styles.bone, styles.lineLg)} />
              <div className={clsx(styles.bone, styles.lineMd)} />
              <div className={clsx(styles.bone, styles.linePrice)} />
              <div className={clsx(styles.cardActions)}>
                <div className={clsx(styles.bone, styles.btn)} />
                <div className={clsx(styles.bone, styles.btn)} />
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div aria-busy aria-label="Загрузка каталога" className={clsx(styles.tableWrap)} role="status">
      <div className={clsx(styles.tableHead)}>
        <div className={clsx(styles.headThumb)} />
        <div className={clsx(styles.bone, styles.headCell)} />
        <div className={clsx(styles.bone, styles.headCell, styles.headSpec)} />
        <div className={clsx(styles.bone, styles.headCell, styles.headSpec)} />
        <div className={clsx(styles.bone, styles.headCell, styles.headSpec)} />
        <div className={clsx(styles.bone, styles.headCell, styles.headSpec)} />
        <div className={clsx(styles.bone, styles.headCell, styles.headSpec)} />
        <div className={clsx(styles.bone, styles.headCell, styles.headPrice)} />
        <div className={clsx(styles.bone, styles.headCell, styles.headActions)} />
      </div>
      {Array.from({ length: TABLE_ROWS }, (_, row) => (
        <div className={clsx(styles.tableRow)} key={row}>
          <div className={clsx(styles.bone, styles.thumb)} />
          <div className={clsx(styles.info)}>
            <div className={clsx(styles.bone, styles.lineSm)} />
            <div className={clsx(styles.bone, styles.lineLg)} />
          </div>
          <div className={clsx(styles.bone, styles.spec)} />
          <div className={clsx(styles.bone, styles.spec)} />
          <div className={clsx(styles.bone, styles.spec)} />
          <div className={clsx(styles.bone, styles.spec)} />
          <div className={clsx(styles.bone, styles.spec)} />
          <div className={clsx(styles.bone, styles.cellPrice)} />
          <div className={clsx(styles.bone, styles.cellAction)} />
        </div>
      ))}
    </div>
  );
};

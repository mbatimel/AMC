'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';

import styles from '../Home.module.css';

export const HomeContent = (): JSX.Element => {
  return (
    <div className={clsx(styles.root)}>
      <Button variant="primary">AMC</Button>
    </div>
  );
};

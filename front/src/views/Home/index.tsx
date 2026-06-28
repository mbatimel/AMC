'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';

import styles from './Home.module.css';

export const Home = (): JSX.Element => {
  return (
    <main className={clsx(styles.root)}>
      <Button variant="primary">AMC</Button>
    </main>
  );
};

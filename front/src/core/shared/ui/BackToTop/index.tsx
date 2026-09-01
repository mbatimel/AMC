'use client';

import clsx from 'clsx';
import { useEffect, useState } from 'react';

import { IconChevronDown } from '@/core/shared/icons';

import styles from './BackToTop.module.css';

const SHOW_AFTER_PX = 400;

export const BackToTop = (): JSX.Element => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const onScroll = (): void => {
      setIsVisible(window.scrollY > SHOW_AFTER_PX);
    };

    onScroll();
    window.addEventListener('scroll', onScroll, { passive: true });

    return () => {
      window.removeEventListener('scroll', onScroll);
    };
  }, []);

  return (
    <button
      aria-label="Вернуться наверх"
      className={clsx(styles.button, !isVisible && styles.buttonHidden)}
      onClick={() => {
        window.scrollTo({ behavior: 'smooth', top: 0 });
      }}
      type="button"
    >
      <IconChevronDown className={clsx(styles.icon)} height={18} width={18} />
    </button>
  );
};

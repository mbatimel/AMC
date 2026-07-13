'use client';

import { useCallback, useEffect, useId, useRef, useState } from 'react';

import { getMobileLayoutMediaQuery } from '@/core/shared/lib/breakpoints';

const FOCUSABLE_SELECTOR =
  'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])';

const getFocusableElements = (container: HTMLElement): HTMLElement[] =>
  Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (element) => !element.hasAttribute('disabled') && element.tabIndex !== -1,
  );

export const useMobileMenu = () => {
  const [isOpen, setIsOpen] = useState(false);
  const drawerId = useId();
  const menuTriggerRef = useRef<HTMLButtonElement>(null);
  const drawerRef = useRef<HTMLElement>(null);

  const openMenu = useCallback((): void => {
    setIsOpen(true);
  }, []);

  const closeMenu = useCallback((): void => {
    setIsOpen(false);
  }, []);

  const toggleMenu = useCallback((): void => {
    setIsOpen((isMenuOpen) => !isMenuOpen);
  }, []);

  useEffect(() => {
    if (!isOpen) {
      return undefined;
    }

    const drawer = drawerRef.current;
    const menuTrigger = menuTriggerRef.current;

    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === 'Escape') {
        closeMenu();
        return;
      }

      if (event.key !== 'Tab' || !drawer) {
        return;
      }

      const focusableElements = getFocusableElements(drawer);

      if (focusableElements.length === 0) {
        return;
      }

      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];
      const activeElement = document.activeElement;

      if (event.shiftKey && activeElement === firstElement) {
        event.preventDefault();
        lastElement.focus();
        return;
      }

      if (!event.shiftKey && activeElement === lastElement) {
        event.preventDefault();
        firstElement.focus();
      }
    };

    document.addEventListener('keydown', onKeyDown);
    document.body.classList.add('overlaypopup-noscroll');

    const [firstFocusableElement] = drawer ? getFocusableElements(drawer) : [];
    firstFocusableElement?.focus();

    return () => {
      document.removeEventListener('keydown', onKeyDown);
      document.body.classList.remove('overlaypopup-noscroll');
      menuTrigger?.focus();
    };
  }, [closeMenu, isOpen]);

  useEffect(() => {
    if (!isOpen) {
      return undefined;
    }

    const mediaQueryList = window.matchMedia(getMobileLayoutMediaQuery());
    const onChange = (): void => {
      if (!mediaQueryList.matches) {
        closeMenu();
      }
    };

    mediaQueryList.addEventListener('change', onChange);

    return () => {
      mediaQueryList.removeEventListener('change', onChange);
    };
  }, [closeMenu, isOpen]);

  return {
    closeMenu,
    drawerId,
    drawerRef,
    isOpen,
    menuTriggerRef,
    openMenu,
    toggleMenu,
  };
};

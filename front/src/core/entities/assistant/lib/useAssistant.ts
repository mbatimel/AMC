'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import {
  $assistantHistory,
  $isAssistantOpen,
  $isAssistantPending,
  $isOperatorMode,
  $isOperatorPending,
  assistantClosed,
  assistantHydrated,
  assistantMessageSent,
  assistantOpened,
  assistantOperatorRequested,
  assistantReset,
  assistantToggled,
} from '../model';

export const useAssistant = () => {
  const [
    history,
    isOpen,
    isPending,
    isOperatorMode,
    isOperatorPending,
    hydrate,
    open,
    close,
    toggle,
    send,
    requestOperator,
    reset,
  ] = useUnit([
    $assistantHistory,
    $isAssistantOpen,
    $isAssistantPending,
    $isOperatorMode,
    $isOperatorPending,
    assistantHydrated,
    assistantOpened,
    assistantClosed,
    assistantToggled,
    assistantMessageSent,
    assistantOperatorRequested,
    assistantReset,
  ]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return {
    close,
    history,
    isOpen,
    isOperatorMode,
    isOperatorPending,
    isPending,
    open,
    requestOperator,
    reset,
    send,
    toggle,
  };
};

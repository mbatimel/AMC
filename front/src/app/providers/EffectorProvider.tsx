'use client';

import { fork } from 'effector';
import { Provider, useUnit } from 'effector-react';
import { useEffect, useState } from 'react';

import { appStarted } from '@/core/app/model';

type EffectorProviderProps = {
  children: React.ReactNode;
};

export const EffectorProvider = ({ children }: EffectorProviderProps): JSX.Element => {
  const [scope] = useState(() => fork());
  const startApp = useUnit(appStarted);

  useEffect(() => {
    startApp();
  }, [startApp]);

  return <Provider value={scope}>{children}</Provider>;
};

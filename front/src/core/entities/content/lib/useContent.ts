'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import {
  $banners,
  $content,
  $contentError,
  $isContentPending,
  $legalDocs,
  contentRequested,
} from '../model';

export const useContent = () => {
  const [content, banners, legalDocs, isPending, error, request] = useUnit([
    $content,
    $banners,
    $legalDocs,
    $isContentPending,
    $contentError,
    contentRequested,
  ]);

  useEffect(() => {
    request();
  }, [request]);

  return { banners, content, error, isPending, legalDocs };
};

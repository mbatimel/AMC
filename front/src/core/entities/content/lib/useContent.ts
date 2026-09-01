'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import {
  $banners,
  $certificates,
  $content,
  $contentError,
  $isContentPending,
  $legalDocs,
  contentRequested,
} from '../model';

export const useContent = () => {
  const [content, banners, certificates, legalDocs, isPending, error, request] = useUnit([
    $content,
    $banners,
    $certificates,
    $legalDocs,
    $isContentPending,
    $contentError,
    contentRequested,
  ]);

  useEffect(() => {
    request();
  }, [request]);

  return { banners, certificates, content, error, isPending, legalDocs };
};

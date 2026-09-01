'use client';

import clsx from 'clsx';
import Link from 'next/link';

import { useContent } from '@/core/entities/content';
import { AppPath, getLegalDocPath } from '@/core/shared/router/paths';
import { InfoCard, InfoPage, InfoPageSkeleton } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import styles from './Legal.module.css';

type LegalProps = {
  docId?: string;
};

const formatDocDate = (value: string): string => {
  if (!value) {
    return '—';
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(date);
};

const FileLink = ({ href, label }: { href: string; label: string }): JSX.Element => (
  <a className={clsx(styles.link)} href={href} rel="noopener noreferrer" target="_blank">
    {label}
  </a>
);

export const Legal = ({ docId }: LegalProps): JSX.Element => {
  const { error, isPending, legalDocs } = useContent();
  const doc = docId ? legalDocs.find((item) => item.id === docId) : undefined;

  if (!docId) {
    return (
      <Page>
        <InfoPage
          description="Оферта, политика конфиденциальности и другие соглашения в актуальной редакции."
          eyebrow="Документы"
          title="Документы и соглашения"
        >
          <InfoCard>
            {isPending && legalDocs.length === 0 ? <InfoPageSkeleton /> : null}
            {error && legalDocs.length === 0 ? <p className={clsx(styles.error)}>{error}</p> : null}
            {!isPending && legalDocs.length === 0 && !error ? (
              <p className={clsx(styles.error)}>Документы ещё не опубликованы.</p>
            ) : null}
            {legalDocs.length > 0 ? (
              <table className={clsx(styles.table)}>
                <thead>
                  <tr>
                    <th>Документ</th>
                    <th>Версия</th>
                    <th>Обновлён</th>
                    <th>Файл</th>
                  </tr>
                </thead>
                <tbody>
                  {legalDocs.map((item) => (
                    <tr key={item.id}>
                      <td>
                        <Link className={clsx(styles.link)} href={getLegalDocPath(item.id)}>
                          {item.name}
                        </Link>
                      </td>
                      <td>v{item.current_version}</td>
                      <td>{formatDocDate(item.updated_at)}</td>
                      <td>
                        {item.file_url ? <FileLink href={item.file_url} label="Скачать" /> : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : null}
          </InfoCard>
        </InfoPage>
      </Page>
    );
  }

  return (
    <Page>
      <InfoPage
        description={
          doc
            ? `Действующая редакция v${doc.current_version}`
            : 'Юридические документы и соглашения портала'
        }
        eyebrow="Документы"
        title={doc?.name ?? 'Документ'}
      >
        <InfoCard>
          {isPending && !doc ? <InfoPageSkeleton /> : null}
          {error && !doc ? <p className={clsx(styles.error)}>{error}</p> : null}
          {!isPending && !doc && !error ? (
            <p className={clsx(styles.error)}>Документ не найден или ещё не опубликован.</p>
          ) : null}
          {doc ? (
            <div className={clsx(styles.detail)}>
              <p className={clsx(styles.meta)}>Обновлён {formatDocDate(doc.updated_at)}</p>
              {doc.file_url ? (
                <FileLink href={doc.file_url} label="Скачать текущую версию" />
              ) : (
                <p className={clsx(styles.error)}>Файл документа пока не загружен.</p>
              )}
            </div>
          ) : null}
        </InfoCard>

        {legalDocs.length > 1 ? (
          <InfoCard title="Другие документы">
            <ul className={clsx(styles.otherDocs)}>
              {legalDocs
                .filter((item) => item.id !== docId)
                .map((item) => (
                  <li key={item.id}>
                    <Link className={clsx(styles.link)} href={getLegalDocPath(item.id)}>
                      {item.name}
                    </Link>
                  </li>
                ))}
              <li>
                <Link className={clsx(styles.link)} href={AppPath.Legal}>
                  Все документы
                </Link>
              </li>
            </ul>
          </InfoCard>
        ) : null}
      </InfoPage>
    </Page>
  );
};

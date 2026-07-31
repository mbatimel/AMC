'use client';

import clsx from 'clsx';
import Link from 'next/link';

import { useContent } from '@/core/entities/content';
import { getLegalDocPath } from '@/core/shared/router/paths';
import { InfoCard, InfoPage, InfoPageSkeleton, InfoText } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import styles from './Legal.module.css';

type LegalProps = {
  docId: string;
};

export const Legal = ({ docId }: LegalProps): JSX.Element => {
  const { error, isPending, legalDocs } = useContent();
  const doc = legalDocs.find((item) => item.id === docId);

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
          {doc ? <InfoText text={doc.body} /> : null}
        </InfoCard>

        {doc && doc.versions.length > 0 ? (
          <InfoCard title="История версий">
            <table className={clsx(styles.table)}>
              <thead>
                <tr>
                  <th>Версия</th>
                  <th>Дата</th>
                  <th>Изменения</th>
                  <th>Автор</th>
                </tr>
              </thead>
              <tbody>
                {doc.versions.map((version) => (
                  <tr key={version.version}>
                    <td>v{version.version}</td>
                    <td>{version.date}</td>
                    <td>{version.summary}</td>
                    <td>{version.author}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </InfoCard>
        ) : null}

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
            </ul>
          </InfoCard>
        ) : null}
      </InfoPage>
    </Page>
  );
};

'use client';

import clsx from 'clsx';

import { useContent } from '@/core/entities/content';
import { InfoCard, InfoPage, InfoPageSkeleton, InfoText } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import styles from './Certificates.module.css';

export const Certificates = (): JSX.Element => {
  const { certificates, content, error, isPending } = useContent();
  const intro = content?.certificates;
  const showIntroSkeleton = isPending && !intro;

  return (
    <Page>
      <InfoPage title={intro?.title || undefined}>
        {showIntroSkeleton ? (
          <InfoCard>
            <InfoPageSkeleton />
          </InfoCard>
        ) : null}

        {intro?.text ? (
          <InfoCard>
            <InfoText text={intro.text} />
          </InfoCard>
        ) : null}

        <InfoCard title={certificates.length > 0 ? 'Перечень документов' : undefined}>
          {isPending && certificates.length === 0 ? <InfoPageSkeleton /> : null}
          {error && certificates.length === 0 ? (
            <p className={clsx(styles.error)}>{error}</p>
          ) : null}
          {!isPending && certificates.length === 0 && !error ? (
            <p className={clsx(styles.empty)}>Сертификаты ещё не опубликованы.</p>
          ) : null}
          {certificates.length > 0 ? (
            <ul className={clsx(styles.list)}>
              {certificates.map((item) => (
                <li className={clsx(styles.item)} key={item.id}>
                  <span>{item.title}</span>
                  {item.file_url ? (
                    <a
                      className={clsx(styles.link)}
                      href={item.file_url}
                      rel="noopener noreferrer"
                      target="_blank"
                    >
                      Скачать
                    </a>
                  ) : (
                    <span className={clsx(styles.empty)}>Файл не загружен</span>
                  )}
                </li>
              ))}
            </ul>
          ) : null}
        </InfoCard>
      </InfoPage>
    </Page>
  );
};

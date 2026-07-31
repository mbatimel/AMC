'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useRouter } from 'next/navigation';

import { useContent } from '@/core/entities/content';
import { AppPath } from '@/core/shared/router/paths';
import { InfoCard, InfoPage, InfoPageSkeleton } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import styles from './Contacts.module.css';

export const Contacts = (): JSX.Element => {
  const router = useRouter();
  const { content, error, isPending } = useContent();
  const contacts = content?.contacts;

  return (
    <Page>
      <InfoPage
        description="Отдел продаж, техническая поддержка и реквизиты предприятия."
        eyebrow="Связь"
        title={contacts?.title ?? 'Контакты'}
      >
        <InfoCard>
          {isPending && !contacts ? <InfoPageSkeleton /> : null}
          {error && !contacts ? <p>{error}</p> : null}

          {contacts ? (
            <dl className={clsx(styles.grid)}>
              <div className={clsx(styles.row)}>
                <dt className={clsx(styles.term)}>Адрес</dt>
                <dd className={clsx(styles.value)}>{contacts.address}</dd>
              </div>
              <div className={clsx(styles.row)}>
                <dt className={clsx(styles.term)}>Телефон</dt>
                <dd className={clsx(styles.value)}>
                  <a href={`tel:${contacts.phone.replace(/[^\d+]/g, '')}`}>{contacts.phone}</a>
                </dd>
              </div>
              <div className={clsx(styles.row)}>
                <dt className={clsx(styles.term)}>E-mail</dt>
                <dd className={clsx(styles.value)}>
                  <a href={`mailto:${contacts.email}`}>{contacts.email}</a>
                </dd>
              </div>
              <div className={clsx(styles.row)}>
                <dt className={clsx(styles.term)}>Часы работы</dt>
                <dd className={clsx(styles.value)}>{contacts.work_hours}</dd>
              </div>
              <div className={clsx(styles.row)}>
                <dt className={clsx(styles.term)}>Реквизиты</dt>
                <dd className={clsx(styles.value)}>{contacts.requisites}</dd>
              </div>
            </dl>
          ) : null}
        </InfoCard>

        {contacts && contacts.managers.length > 0 ? (
          <InfoCard title="Кому писать">
            <div className={clsx(styles.managers)}>
              {contacts.managers.map((manager) => (
                <article className={clsx(styles.manager)} key={manager.email}>
                  <h3 className={clsx(styles.managerName)}>{manager.name}</h3>
                  <p className={clsx(styles.managerRole)}>{manager.role}</p>
                  <p className={clsx(styles.managerContact)}>
                    <a href={`tel:${manager.phone.replace(/[^\d+]/g, '')}`}>{manager.phone}</a>
                  </p>
                  <p className={clsx(styles.managerContact)}>
                    <a href={`mailto:${manager.email}`}>{manager.email}</a>
                  </p>
                </article>
              ))}
            </div>
          </InfoCard>
        ) : null}

        <InfoCard>
          <Button onPress={() => router.push(AppPath.Support)} variant="primary">
            Написать в поддержку
          </Button>
        </InfoCard>
      </InfoPage>
    </Page>
  );
};

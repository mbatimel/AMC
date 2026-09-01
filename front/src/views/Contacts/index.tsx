'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';

import { useContent } from '@/core/entities/content';
import {
  IconBuilding,
  IconCalendar,
  IconEmail,
  IconLocation,
  IconMail,
  IconPhone,
  IconSupport,
} from '@/core/shared/icons';
import { Page } from '@/core/shared/ui/Page';
import { toastShown } from '@/core/shared/ui/Toast/model';

import styles from './Contacts.module.css';
import {
  downloadRequisitesFile,
  formatRequisitesText,
  isYandexMapEmbedUrl,
  toTelHref,
} from './lib/contactsData';

const departmentIcons = [IconSupport, IconPhone, IconBuilding] as const;

export const Contacts = (): JSX.Element => {
  const showToast = useUnit(toastShown);
  const { content, error, isPending } = useContent();
  const contacts = content?.contacts;
  const title = contacts?.title ?? 'Контакты';
  const subtitle = contacts?.subtitle ?? 'Контакты отделов и представительств компании.';
  const workHours = contacts?.work_hours ?? '';
  const managers = contacts?.managers ?? [];
  const offices = contacts?.offices ?? [];
  const requisiteItems = contacts?.requisite_items ?? [];
  const warehouseCaption = contacts?.warehouse_map_caption ?? '';
  const warehouseMapUrl = contacts?.warehouse_map_url ?? '';
  const hasWarehouseMap = Boolean(warehouseMapUrl && isYandexMapEmbedUrl(warehouseMapUrl));

  const copyRequisites = async (): Promise<void> => {
    const text = formatRequisitesText(requisiteItems);

    try {
      await navigator.clipboard.writeText(text);
      showToast({ message: 'Реквизиты скопированы', tone: 'success' });
    } catch {
      showToast({ message: 'Не удалось скопировать реквизиты', tone: 'error' });
    }
  };

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <section className={clsx(styles.hero)}>
          <div className={clsx(styles.heroInner)}>
            {workHours ? <p className={clsx(styles.heroBadge)}>Мы на связи {workHours}</p> : null}
            <h1 className={clsx(styles.heroTitle)}>{title}</h1>
            {subtitle ? <p className={clsx(styles.heroDescription)}>{subtitle}</p> : null}
          </div>
        </section>

        <div className={clsx(styles.container)}>
          {isPending && !contacts ? <p className={clsx(styles.status)}>Загрузка…</p> : null}
          {error && !contacts ? <p className={clsx(styles.error)}>{error}</p> : null}

          {managers.length > 0 ? (
            <section aria-label="Отделы" className={clsx(styles.departments)}>
              {managers.map((manager, index) => {
                const Icon = departmentIcons[index] ?? IconBuilding;
                const isFeatured = index === 0;

                return (
                  <article
                    className={clsx(styles.departmentCard, isFeatured && styles.departmentFeatured)}
                    key={`${manager.name}-${manager.email}-${index}`}
                  >
                    <div aria-hidden className={clsx(styles.departmentIcon)}>
                      <Icon currentColor="currentColor" height={22} width={22} />
                    </div>

                    <h2 className={clsx(styles.departmentTitle)}>{manager.name}</h2>

                    <ul className={clsx(styles.departmentMeta)}>
                      {manager.email ? (
                        <li>
                          <IconEmail currentColor="currentColor" height={14} width={14} />
                          <a href={`mailto:${manager.email}`}>{manager.email}</a>
                        </li>
                      ) : null}
                      {manager.phone ? (
                        <li>
                          <IconPhone currentColor="currentColor" height={14} width={14} />
                          <a href={toTelHref(manager.phone)}>{manager.phone}</a>
                        </li>
                      ) : null}
                      {manager.role ? (
                        <li>
                          {index === 1 ? (
                            <IconCalendar currentColor="currentColor" height={14} width={14} />
                          ) : (
                            <IconMail currentColor="currentColor" height={14} width={14} />
                          )}
                          <span>{manager.role}</span>
                        </li>
                      ) : null}
                    </ul>

                    {isFeatured ? (
                      <div className={clsx(styles.departmentActions)}>
                        {manager.email ? (
                          <a
                            className={clsx(styles.primaryButton)}
                            href={`mailto:${manager.email}`}
                          >
                            Письмо
                          </a>
                        ) : null}
                        {manager.phone ? (
                          <a
                            className={clsx(styles.secondaryButton)}
                            href={toTelHref(manager.phone)}
                          >
                            Позвонить
                          </a>
                        ) : null}
                      </div>
                    ) : null}
                  </article>
                );
              })}
            </section>
          ) : null}

          {offices.length > 0 ? (
            <section aria-labelledby="offices-title" className={clsx(styles.officesSection)}>
              <div className={clsx(styles.sectionIntro)}>
                <p className={clsx(styles.sectionBadge)}>География</p>
                <h2 className={clsx(styles.sectionTitle)} id="offices-title">
                  Представительства
                </h2>
              </div>

              <div className={clsx(styles.officesGrid)}>
                {offices.map((office) => (
                  <article className={clsx(styles.officeCard)} key={office.city}>
                    <div className={clsx(styles.officeHeader)}>
                      <span aria-hidden className={clsx(styles.officePin)}>
                        <IconLocation currentColor="currentColor" height={16} width={16} />
                      </span>
                      <h3 className={clsx(styles.officeCity)}>{office.city}</h3>
                      {office.is_main ? (
                        <span className={clsx(styles.officeMain)}>главный офис</span>
                      ) : null}
                    </div>
                    <ul className={clsx(styles.officeMeta)}>
                      {office.address ? (
                        <li>
                          <IconLocation currentColor="currentColor" height={13} width={13} />
                          <span>{office.address}</span>
                        </li>
                      ) : null}
                      {office.phone ? (
                        <li>
                          <IconPhone currentColor="currentColor" height={13} width={13} />
                          <a href={toTelHref(office.phone)}>{office.phone}</a>
                        </li>
                      ) : null}
                      {office.email ? (
                        <li>
                          <IconEmail currentColor="currentColor" height={13} width={13} />
                          <a href={`mailto:${office.email}`}>{office.email}</a>
                        </li>
                      ) : null}
                    </ul>
                  </article>
                ))}
              </div>
            </section>
          ) : null}

          {hasWarehouseMap || warehouseCaption ? (
            <section aria-labelledby="warehouse-map-title" className={clsx(styles.mapSection)}>
              <div className={clsx(styles.sectionIntro)}>
                <p className={clsx(styles.sectionBadge)}>Склад</p>
                <h2 className={clsx(styles.sectionTitle)} id="warehouse-map-title">
                  Схема проезда на склад
                </h2>
                {warehouseCaption ? (
                  <p className={clsx(styles.mapCaption)}>{warehouseCaption}</p>
                ) : null}
              </div>

              {hasWarehouseMap ? (
                <div className={clsx(styles.mapFrame)}>
                  <iframe
                    allowFullScreen
                    className={clsx(styles.mapIframe)}
                    src={warehouseMapUrl}
                    title="Схема проезда на склад"
                  />
                </div>
              ) : null}
            </section>
          ) : null}

          {requisiteItems.length > 0 ? (
            <section aria-labelledby="requisites-title" className={clsx(styles.requisites)}>
              <header className={clsx(styles.requisitesHeader)}>
                <span aria-hidden className={clsx(styles.requisitesIcon)}>
                  <IconBuilding currentColor="currentColor" height={18} width={18} />
                </span>
                <h2 className={clsx(styles.requisitesTitle)} id="requisites-title">
                  Реквизиты компании
                </h2>
                <div className={clsx(styles.requisitesActions)}>
                  <Button onPress={() => void copyRequisites()} size="sm" variant="secondary">
                    Скопировать реквизиты в буфер обмена
                  </Button>
                  <Button
                    onPress={() =>
                      downloadRequisitesFile(formatRequisitesText(requisiteItems), 'requisites.txt')
                    }
                    size="sm"
                    variant="secondary"
                  >
                    Скачать файл с реквизитами
                  </Button>
                </div>
              </header>

              <dl className={clsx(styles.requisitesList)}>
                {requisiteItems.map((row) => (
                  <div className={clsx(styles.requisitesRow)} key={row.label}>
                    <dt>{row.label}</dt>
                    <dd>{row.value}</dd>
                  </div>
                ))}
              </dl>
            </section>
          ) : null}
        </div>
      </div>
    </Page>
  );
};

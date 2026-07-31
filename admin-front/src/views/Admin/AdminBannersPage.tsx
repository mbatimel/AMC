'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect, useState } from 'react';

import type { BannerItem, BannersSettings } from '@/core/shared/api/content';

import { useContent } from '@/core/entities/content';

import styles from './Admin.module.css';
import { $contentSaveError, $isBannersSaving, bannersSaveRequested } from './model/content';
import { AdminPageHeader } from './ui/AdminPageHeader';

const createBanner = (): BannerItem => ({
  dateFrom: '',
  dateTo: '',
  id: `banner-${Date.now().toString(36)}`,
  image: '',
  is_active: true,
  link: '',
  sort_order: 0,
  subtitle: '',
  title: 'Новый баннер',
});

export const AdminBannersPage = (): JSX.Element => {
  const { banners } = useContent();
  const [isSaving, error, save] = useUnit([
    $isBannersSaving,
    $contentSaveError,
    bannersSaveRequested,
  ]);
  const [draft, setDraft] = useState<BannersSettings | null>(null);

  useEffect(() => {
    if (banners) {
      setDraft({ delay_sec: banners.delay_sec, items: banners.items.map((item) => ({ ...item })) });
    }
  }, [banners]);

  if (!draft) {
    return (
      <>
        <AdminPageHeader title="Баннеры" />
        <p className={clsx(styles.hint)}>Загружаем баннеры…</p>
      </>
    );
  }

  const patchItem = (index: number, patch: Partial<BannerItem>): void => {
    setDraft((previous) =>
      previous
        ? {
            ...previous,
            items: previous.items.map((item, itemIndex) =>
              itemIndex === index ? { ...item, ...patch } : item,
            ),
          }
        : previous,
    );
  };

  const moveItem = (index: number, direction: -1 | 1): void => {
    setDraft((previous) => {
      if (!previous) {
        return previous;
      }

      const target = index + direction;

      if (target < 0 || target >= previous.items.length) {
        return previous;
      }

      const items = [...previous.items];
      const [moved] = items.splice(index, 1);

      items.splice(target, 0, moved);

      return { ...previous, items };
    });
  };

  return (
    <>
      <AdminPageHeader
        actions={
          <>
            <button
              className={clsx(styles.smallButton)}
              onClick={() =>
                setDraft((previous) =>
                  previous ? { ...previous, items: [...previous.items, createBanner()] } : previous,
                )
              }
              type="button"
            >
              Добавить баннер
            </button>
            <Button isDisabled={isSaving} onPress={() => save(draft)} variant="primary">
              {isSaving ? 'Сохраняем…' : 'Сохранить'}
            </Button>
          </>
        }
        subtitle="Промо-блоки на главной странице и в разделе «Акции»"
        title="Баннеры"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <section className={clsx(styles.card)}>
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="banner-delay">
            Интервал смены, сек
          </label>
          <input
            className={clsx(styles.input)}
            id="banner-delay"
            min={1}
            onChange={(event) =>
              setDraft((previous) =>
                previous ? { ...previous, delay_sec: Number(event.target.value) || 1 } : previous,
              )
            }
            type="number"
            value={draft.delay_sec}
          />
        </div>
      </section>

      <div className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Список баннеров ({draft.items.length})</h2>
        <div className={clsx(styles.listEditor)}>
          {draft.items.map((item, index) => (
            <article className={clsx(styles.bannerItem)} key={item.id}>
              <div className={clsx(styles.bannerHeader)}>
                <strong>#{index + 1}</strong>
                <div className={clsx(styles.rowActions)}>
                  <button
                    className={clsx(styles.smallButton)}
                    onClick={() => moveItem(index, -1)}
                    type="button"
                  >
                    ↑
                  </button>
                  <button
                    className={clsx(styles.smallButton)}
                    onClick={() => moveItem(index, 1)}
                    type="button"
                  >
                    ↓
                  </button>
                  <button
                    className={clsx(styles.smallButton, styles.smallButtonDanger)}
                    onClick={() =>
                      setDraft((previous) =>
                        previous
                          ? {
                              ...previous,
                              items: previous.items.filter((_, i) => i !== index),
                            }
                          : previous,
                      )
                    }
                    type="button"
                  >
                    Удалить
                  </button>
                </div>
              </div>

              <div className={clsx(styles.formGrid)}>
                <div className={clsx(styles.field)}>
                  <label className={clsx(styles.label)} htmlFor={`banner-title-${item.id}`}>
                    Заголовок
                  </label>
                  <input
                    className={clsx(styles.input)}
                    id={`banner-title-${item.id}`}
                    onChange={(event) => patchItem(index, { title: event.target.value })}
                    value={item.title}
                  />
                </div>
                <div className={clsx(styles.field)}>
                  <label className={clsx(styles.label)} htmlFor={`banner-subtitle-${item.id}`}>
                    Подзаголовок
                  </label>
                  <input
                    className={clsx(styles.input)}
                    id={`banner-subtitle-${item.id}`}
                    onChange={(event) => patchItem(index, { subtitle: event.target.value })}
                    value={item.subtitle}
                  />
                </div>
                <div className={clsx(styles.field)}>
                  <label className={clsx(styles.label)} htmlFor={`banner-link-${item.id}`}>
                    Ссылка
                  </label>
                  <input
                    className={clsx(styles.input)}
                    id={`banner-link-${item.id}`}
                    onChange={(event) => patchItem(index, { link: event.target.value })}
                    placeholder="/catalog?q=сверло"
                    value={item.link}
                  />
                </div>
                <div className={clsx(styles.field)}>
                  <label className={clsx(styles.label)} htmlFor={`banner-image-${item.id}`}>
                    URL изображения
                  </label>
                  <input
                    className={clsx(styles.input)}
                    id={`banner-image-${item.id}`}
                    onChange={(event) => patchItem(index, { image: event.target.value })}
                    value={item.image}
                  />
                </div>
                <div className={clsx(styles.field)}>
                  <label className={clsx(styles.label)} htmlFor={`banner-from-${item.id}`}>
                    Показывать с
                  </label>
                  <input
                    className={clsx(styles.input)}
                    id={`banner-from-${item.id}`}
                    onChange={(event) => patchItem(index, { dateFrom: event.target.value })}
                    type="date"
                    value={item.dateFrom}
                  />
                </div>
                <div className={clsx(styles.field)}>
                  <label className={clsx(styles.label)} htmlFor={`banner-to-${item.id}`}>
                    Показывать по
                  </label>
                  <input
                    className={clsx(styles.input)}
                    id={`banner-to-${item.id}`}
                    onChange={(event) => patchItem(index, { dateTo: event.target.value })}
                    type="date"
                    value={item.dateTo}
                  />
                </div>
              </div>

              <label className={clsx(styles.checkboxRow)}>
                <input
                  checked={item.is_active}
                  onChange={(event) => patchItem(index, { is_active: event.target.checked })}
                  type="checkbox"
                />
                Показывать на сайте
              </label>
            </article>
          ))}
        </div>
      </div>
    </>
  );
};

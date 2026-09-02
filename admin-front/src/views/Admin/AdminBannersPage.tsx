'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useCallback, useEffect, useRef, useState } from 'react';

import type { BannerItem, BannersSettings } from '@/core/shared/api/content';

import { $adminUserId } from '@/core/entities/adminSession';
import {
  createBannerRequest,
  deleteBannerRequest,
  fetchAdminBannersRequest,
  updateBannerDelayRequest,
  updateBannerRequest,
} from '@/core/shared/api/content';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { createClientId } from '@/core/shared/lib/createClientId';

import styles from './Admin.module.css';
import bannerStyles from './AdminBanners.module.css';
import { AdminPageHeader } from './ui/AdminPageHeader';
import { FileThumb } from './ui/FileThumb';

const createBanner = (): BannerItem => ({
  dateFrom: '',
  dateTo: '',
  id: createClientId(),
  image: '',
  is_active: true,
  link: '',
  sort_order: 0,
  subtitle: '',
  title: 'Новый баннер',
});

export const AdminBannersPage = (): JSX.Element => {
  const adminUserId = useUnit($adminUserId);
  const [draft, setDraft] = useState<BannersSettings | null>(null);
  const [persistedIds, setPersistedIds] = useState<string[]>([]);
  const [files, setFiles] = useState<Record<string, File | undefined>>({});
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<null | string>(null);
  const [addedId, setAddedId] = useState<null | string>(null);
  const addedBannerRef = useRef<HTMLElement | null>(null);

  const load = useCallback(async (): Promise<void> => {
    if (!adminUserId) {
      return;
    }
    try {
      const banners = await fetchAdminBannersRequest(adminUserId);

      setDraft({
        ...banners,
        items: Array.isArray(banners.items) ? banners.items.map((item) => ({ ...item })) : [],
      });
      setPersistedIds(Array.isArray(banners.items) ? banners.items.map((item) => item.id) : []);
      setFiles({});
      setError(null);
    } catch (loadError) {
      setError(toDisplayErrorMessage(loadError, 'Не удалось загрузить баннеры'));
    }
  }, [adminUserId]);

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0);

    return () => window.clearTimeout(timeout);
  }, [load]);

  const addBanner = (): void => {
    const banner = createBanner();

    setDraft((previous) =>
      previous ? { ...previous, items: [banner, ...(previous.items ?? [])] } : previous,
    );
    setAddedId(banner.id);
  };

  useEffect(() => {
    if (!addedId) {
      return;
    }

    addedBannerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }, [addedId]);

  if (!draft) {
    return (
      <>
        <AdminPageHeader title="Баннеры" />
        <p className={clsx(error ? styles.error : styles.hint)}>{error ?? 'Загружаем баннеры…'}</p>
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
      if (!previous) return previous;
      const target = index + direction;

      if (target < 0 || target >= previous.items.length) return previous;
      const items = [...previous.items];
      const [moved] = items.splice(index, 1);

      items.splice(target, 0, moved);
      return { ...previous, items };
    });
  };

  const save = async (): Promise<void> => {
    if (!adminUserId) return;
    setIsSaving(true);
    setError(null);
    try {
      const currentIds = new Set(draft.items.map((item) => item.id));

      for (const id of persistedIds) {
        if (!currentIds.has(id)) await deleteBannerRequest(adminUserId, id);
      }
      for (const [index, rawItem] of draft.items.entries()) {
        const item = { ...rawItem, sort_order: index + 1 };
        const file = files[item.id];

        if (item.id.startsWith('new-')) {
          if (!file) throw new Error(`Выберите изображение для баннера «${item.title}»`);
          await createBannerRequest(adminUserId, item, file);
        } else {
          await updateBannerRequest(adminUserId, item, file);
        }
      }
      await updateBannerDelayRequest(adminUserId, draft.delay_sec);
      await load();
    } catch (saveError) {
      setError(toDisplayErrorMessage(saveError, 'Не удалось сохранить баннеры'));
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <>
      <AdminPageHeader
        actions={
          <>
            <button className={clsx(styles.smallButton)} onClick={addBanner} type="button">
              Добавить баннер
            </button>
            <Button isDisabled={isSaving} onPress={() => void save()} variant="primary">
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
            onChange={(event) => setDraft({ ...draft, delay_sec: Number(event.target.value) || 1 })}
            type="number"
            value={draft.delay_sec}
          />
        </div>
      </section>

      <div className={clsx(styles.card)}>
        <div className={clsx(styles.bannerHeader)}>
          <h2 className={clsx(bannerStyles.listTitle)}>Список баннеров ({draft.items.length})</h2>
          <button className={clsx(styles.smallButton)} onClick={addBanner} type="button">
            Добавить баннер
          </button>
        </div>
        <div className={clsx(styles.listEditor)}>
          {draft.items.length === 0 ? (
            <p className={clsx(styles.hint)}>
              Баннеров пока нет. Добавьте первый — карточка появится здесь.
            </p>
          ) : null}
          {draft.items.map((item, index) => (
            <article
              className={clsx(
                bannerStyles.bannerItem,
                item.id === addedId && bannerStyles.bannerItemNew,
              )}
              key={item.id}
              ref={item.id === addedId ? addedBannerRef : undefined}
            >
              <div className={clsx(styles.bannerHeader)}>
                {' '}
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
                      setDraft({
                        ...draft,
                        items: draft.items.filter((_, itemIndex) => itemIndex !== index),
                      })
                    }
                    type="button"
                  >
                    Удалить
                  </button>
                </div>
              </div>
              {item.image || files[item.id] ? (
                <div className={clsx(bannerStyles.previewRow)}>
                  <FileThumb file={files[item.id]} href={item.image || undefined} />
                  <div className={clsx(bannerStyles.previewMeta)}>
                    {item.image ? (
                      <a href={item.image} rel="noreferrer" target="_blank">
                        {files[item.id] ? 'Открыть текущее изображение' : 'Текущее изображение'}
                      </a>
                    ) : null}
                    {files[item.id] ? (
                      <p className={clsx(bannerStyles.previewName)}>
                        Новый файл: {files[item.id]?.name}
                      </p>
                    ) : null}
                  </div>
                </div>
              ) : null}
              <div className={clsx(styles.formGrid)}>
                <label className={clsx(styles.field)}>
                  Заголовок
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchItem(index, { title: event.target.value })}
                    value={item.title}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Подзаголовок
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchItem(index, { subtitle: event.target.value })}
                    value={item.subtitle}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Ссылка
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchItem(index, { link: event.target.value })}
                    value={item.link}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Изображение
                  <input
                    accept="image/jpeg,image/png,image/webp"
                    onChange={(event) =>
                      setFiles((previous) => ({ ...previous, [item.id]: event.target.files?.[0] }))
                    }
                    type="file"
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Показывать с
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchItem(index, { dateFrom: event.target.value })}
                    type="date"
                    value={item.dateFrom}
                  />
                </label>
                <label className={clsx(styles.field)}>
                  Показывать по
                  <input
                    className={clsx(styles.input)}
                    onChange={(event) => patchItem(index, { dateTo: event.target.value })}
                    type="date"
                    value={item.dateTo}
                  />
                </label>
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

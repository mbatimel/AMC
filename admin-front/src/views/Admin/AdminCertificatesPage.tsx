'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useCallback, useEffect, useState } from 'react';

import type { Certificate } from '@/core/shared/api/certificates';

import { $adminUserId } from '@/core/entities/adminSession';
import {
  createCertificateRequest,
  deleteCertificateRequest,
  fetchAdminCertificatesRequest,
  updateCertificateRequest,
} from '@/core/shared/api/certificates';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { assertDocumentFile, fileToBase64 } from '@/core/shared/lib/fileToBase64';
import { toastShown } from '@/core/shared/ui/Toast/model';

import styles from './Admin.module.css';
import bannerStyles from './AdminBanners.module.css';
import { AdminPageHeader } from './ui/AdminPageHeader';
import { DocumentFileField } from './ui/DocumentFileField';

type CreateDraft = {
  file: File | null;
  isActive: boolean;
  sortOrder: string;
  title: string;
};

type ItemDraft = {
  file: File | null;
  isActive: boolean;
  sortOrder: string;
  title: string;
};

const EMPTY_CREATE: CreateDraft = {
  file: null,
  isActive: true,
  sortOrder: '0',
  title: '',
};

const draftFromCertificate = (item: Certificate): ItemDraft => ({
  file: null,
  isActive: item.is_active,
  sortOrder: String(item.sort_order),
  title: item.title,
});

export const AdminCertificatesPage = (): JSX.Element => {
  const adminUserId = useUnit($adminUserId);
  const showToast = useUnit(toastShown);
  const [items, setItems] = useState<Certificate[]>([]);
  const [drafts, setDrafts] = useState<Record<string, ItemDraft>>({});
  const [createDraft, setCreateDraft] = useState<CreateDraft>(EMPTY_CREATE);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<null | string>(null);

  const load = useCallback(async (): Promise<void> => {
    if (!adminUserId) {
      return;
    }

    setIsLoading(true);
    try {
      const certificates = await fetchAdminCertificatesRequest(adminUserId);

      setItems(certificates);
      setDrafts(
        Object.fromEntries(certificates.map((item) => [item.id, draftFromCertificate(item)])),
      );
      setError(null);
    } catch (loadError) {
      setError(toDisplayErrorMessage(loadError, 'Не удалось загрузить сертификаты'));
    } finally {
      setIsLoading(false);
    }
  }, [adminUserId]);

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0);

    return () => window.clearTimeout(timeout);
  }, [load]);

  const patchDraft = (id: string, patch: Partial<ItemDraft>): void => {
    setDrafts((previous) => ({
      ...previous,
      [id]: { ...(previous[id] ?? EMPTY_CREATE), ...patch },
    }));
  };

  const create = async (): Promise<void> => {
    if (!adminUserId) {
      return;
    }
    if (!createDraft.file) {
      setError('Выберите файл сертификата (PDF, JPG, PNG или WEBP)');
      return;
    }

    setIsSaving(true);
    setError(null);
    try {
      assertDocumentFile(createDraft.file);
      await createCertificateRequest(adminUserId, {
        fileContentBase64: await fileToBase64(createDraft.file),
        fileName: createDraft.file.name,
        isActive: createDraft.isActive,
        sortOrder: Number(createDraft.sortOrder) || 0,
        title: createDraft.title.trim(),
      });
      setCreateDraft(EMPTY_CREATE);
      showToast({ message: 'Сертификат добавлен', tone: 'success' });
      await load();
    } catch (saveError) {
      setError(toDisplayErrorMessage(saveError, 'Не удалось добавить сертификат'));
    } finally {
      setIsSaving(false);
    }
  };

  const saveItem = async (item: Certificate): Promise<void> => {
    if (!adminUserId) {
      return;
    }

    const draft = drafts[item.id] ?? draftFromCertificate(item);

    setIsSaving(true);
    setError(null);
    try {
      if (draft.file) {
        assertDocumentFile(draft.file);
      }
      await updateCertificateRequest(adminUserId, {
        certId: item.id,
        fileContentBase64: draft.file ? await fileToBase64(draft.file) : '',
        fileName: draft.file?.name ?? '',
        isActive: draft.isActive,
        sortOrder: Number(draft.sortOrder) || 0,
        title: draft.title.trim(),
      });
      showToast({ message: 'Сертификат сохранён', tone: 'success' });
      await load();
    } catch (saveError) {
      setError(toDisplayErrorMessage(saveError, 'Не удалось сохранить сертификат'));
    } finally {
      setIsSaving(false);
    }
  };

  const remove = async (certId: string): Promise<void> => {
    if (!adminUserId) {
      return;
    }

    setIsSaving(true);
    setError(null);
    try {
      await deleteCertificateRequest(adminUserId, certId);
      showToast({ message: 'Сертификат удалён', tone: 'success' });
      await load();
    } catch (saveError) {
      setError(toDisplayErrorMessage(saveError, 'Не удалось удалить сертификат'));
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <>
      <AdminPageHeader
        subtitle="Файлы PDF и изображений. Неактивные сертификаты скрыты на витрине"
        title="Сертификаты"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Добавить сертификат</h2>
        <div className={clsx(styles.formGrid)}>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="cert-title">
              Название
            </label>
            <input
              className={clsx(styles.input)}
              id="cert-title"
              onChange={(event) =>
                setCreateDraft((previous) => ({ ...previous, title: event.target.value }))
              }
              placeholder="Сертификат соответствия ГОСТ"
              value={createDraft.title}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="cert-sort">
              Порядок
            </label>
            <input
              className={clsx(styles.input)}
              id="cert-sort"
              min={0}
              onChange={(event) =>
                setCreateDraft((previous) => ({ ...previous, sortOrder: event.target.value }))
              }
              type="number"
              value={createDraft.sortOrder}
            />
          </div>
        </div>
        <DocumentFileField
          file={createDraft.file}
          id="cert-file"
          isDisabled={isSaving}
          isRequired
          label="Файл"
          onChange={(file) => setCreateDraft((previous) => ({ ...previous, file }))}
        />
        <label className={clsx(styles.checkboxRow)}>
          <input
            checked={createDraft.isActive}
            onChange={(event) =>
              setCreateDraft((previous) => ({ ...previous, isActive: event.target.checked }))
            }
            type="checkbox"
          />
          Показывать на сайте
        </label>
        <div className={clsx(styles.actionsRow)}>
          <Button isDisabled={isSaving} onPress={() => void create()} variant="primary">
            {isSaving ? 'Сохраняем…' : 'Добавить'}
          </Button>
        </div>
      </section>

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Перечень</h2>
        {isLoading && items.length === 0 ? (
          <p className={clsx(styles.hint)}>Загружаем сертификаты…</p>
        ) : null}
        {!isLoading && items.length === 0 ? (
          <p className={clsx(styles.hint)}>Сертификатов пока нет</p>
        ) : null}
        <div className={clsx(styles.listEditor)}>
          {items.map((item) => {
            const draft = drafts[item.id] ?? draftFromCertificate(item);

            return (
              <article className={clsx(bannerStyles.bannerItem)} key={item.id}>
                <div className={clsx(styles.formGrid)}>
                  <div className={clsx(styles.field)}>
                    <label className={clsx(styles.label)} htmlFor={`cert-title-${item.id}`}>
                      Название
                    </label>
                    <input
                      className={clsx(styles.input)}
                      id={`cert-title-${item.id}`}
                      onChange={(event) => patchDraft(item.id, { title: event.target.value })}
                      value={draft.title}
                    />
                  </div>
                  <div className={clsx(styles.field)}>
                    <label className={clsx(styles.label)} htmlFor={`cert-sort-${item.id}`}>
                      Порядок
                    </label>
                    <input
                      className={clsx(styles.input)}
                      id={`cert-sort-${item.id}`}
                      min={0}
                      onChange={(event) => patchDraft(item.id, { sortOrder: event.target.value })}
                      type="number"
                      value={draft.sortOrder}
                    />
                  </div>
                </div>
                <DocumentFileField
                  currentFileHref={item.file_url || undefined}
                  currentFileLabel="Открыть текущий файл"
                  file={draft.file}
                  id={`cert-file-${item.id}`}
                  isDisabled={isSaving}
                  label="Заменить файл"
                  onChange={(file) => patchDraft(item.id, { file })}
                />
                <label className={clsx(styles.checkboxRow)}>
                  <input
                    checked={draft.isActive}
                    onChange={(event) => patchDraft(item.id, { isActive: event.target.checked })}
                    type="checkbox"
                  />
                  Показывать на сайте
                </label>
                <div className={clsx(styles.rowActions)}>
                  <Button
                    isDisabled={isSaving}
                    onPress={() => void saveItem(item)}
                    size="sm"
                    variant="primary"
                  >
                    Сохранить
                  </Button>
                  <button
                    className={clsx(styles.smallButton, styles.smallButtonDanger)}
                    disabled={isSaving}
                    onClick={() => void remove(item.id)}
                    type="button"
                  >
                    Удалить
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      </section>
    </>
  );
};

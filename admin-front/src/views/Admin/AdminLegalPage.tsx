'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useCallback, useEffect, useState } from 'react';

import type { LegalDoc, LegalDocVersion } from '@/core/shared/api/legalDocs';

import { $adminUserId } from '@/core/entities/adminSession';
import { contentInvalidated } from '@/core/entities/content';
import {
  createLegalDocRequest,
  deleteLegalDocRequest,
  fetchAdminLegalDocsRequest,
  fetchLegalDocVersionsRequest,
  replaceLegalDocFileRequest,
} from '@/core/shared/api/legalDocs';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import {
  assertDocumentFile,
  DOCUMENT_FILE_HINT,
  fileToBase64,
} from '@/core/shared/lib/fileToBase64';
import { toastShown } from '@/core/shared/ui/Toast/model';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import { AdminPageHeader } from './ui/AdminPageHeader';
import { DocumentFileField } from './ui/DocumentFileField';
import { FileThumb } from './ui/FileThumb';

type CreateDraft = {
  file: File | null;
  id: string;
  name: string;
  summary: string;
  version: string;
};

type ReplaceDraft = {
  file: File | null;
  summary: string;
  version: string;
};

const EMPTY_CREATE: CreateDraft = {
  file: null,
  id: '',
  name: '',
  summary: '',
  version: '1.0',
};

const EMPTY_REPLACE: ReplaceDraft = {
  file: null,
  summary: '',
  version: '',
};

export const AdminLegalPage = (): JSX.Element => {
  const adminUserId = useUnit($adminUserId);
  const invalidateContent = useUnit(contentInvalidated);
  const showToast = useUnit(toastShown);
  const [docs, setDocs] = useState<LegalDoc[]>([]);
  const [versions, setVersions] = useState<LegalDocVersion[]>([]);
  const [activeId, setActiveId] = useState('');
  const [createDraft, setCreateDraft] = useState<CreateDraft>(EMPTY_CREATE);
  const [replaceDraft, setReplaceDraft] = useState<ReplaceDraft>(EMPTY_REPLACE);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<null | string>(null);

  const load = useCallback(async (): Promise<void> => {
    if (!adminUserId) {
      return;
    }

    setIsLoading(true);
    try {
      const items = await fetchAdminLegalDocsRequest(adminUserId);

      setDocs(items);
      setError(null);
      setActiveId((current) => current || items[0]?.id || '');
    } catch (loadError) {
      setError(toDisplayErrorMessage(loadError, 'Не удалось загрузить документы'));
    } finally {
      setIsLoading(false);
    }
  }, [adminUserId]);

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0);

    return () => window.clearTimeout(timeout);
  }, [load]);

  const loadVersions = useCallback(
    async (docId: string): Promise<void> => {
      if (!adminUserId || !docId) {
        setVersions([]);
        return;
      }

      try {
        setVersions(await fetchLegalDocVersionsRequest(adminUserId, docId));
      } catch (loadError) {
        setError(toDisplayErrorMessage(loadError, 'Не удалось загрузить историю версий'));
      }
    },
    [adminUserId],
  );

  useEffect(() => {
    const timeout = window.setTimeout(() => void loadVersions(activeId), 0);

    return () => window.clearTimeout(timeout);
  }, [activeId, loadVersions]);

  const active = docs.find((doc) => doc.id === activeId);

  const create = async (): Promise<void> => {
    if (!adminUserId) {
      return;
    }
    if (!createDraft.file) {
      setError('Выберите файл документа (PDF, JPG, PNG или WEBP)');
      return;
    }

    setIsSaving(true);
    setError(null);
    try {
      assertDocumentFile(createDraft.file);
      await createLegalDocRequest(adminUserId, {
        fileContentBase64: await fileToBase64(createDraft.file),
        fileName: createDraft.file.name,
        id: createDraft.id.trim(),
        name: createDraft.name.trim(),
        summary: createDraft.summary.trim(),
        version: createDraft.version.trim() || '1.0',
      });
      setCreateDraft(EMPTY_CREATE);
      showToast({ message: 'Документ добавлен', tone: 'success' });
      invalidateContent();
      await load();
    } catch (saveError) {
      setError(toDisplayErrorMessage(saveError, 'Не удалось добавить документ'));
    } finally {
      setIsSaving(false);
    }
  };

  const replace = async (): Promise<void> => {
    if (!adminUserId || !active) {
      return;
    }
    if (!replaceDraft.file) {
      setError('Выберите новый файл документа');
      return;
    }

    setIsSaving(true);
    setError(null);
    try {
      assertDocumentFile(replaceDraft.file);
      await replaceLegalDocFileRequest(adminUserId, {
        docId: active.id,
        fileContentBase64: await fileToBase64(replaceDraft.file),
        fileName: replaceDraft.file.name,
        summary: replaceDraft.summary.trim(),
        version: replaceDraft.version.trim() || active.current_version,
      });
      setReplaceDraft(EMPTY_REPLACE);
      showToast({ message: 'Файл документа обновлён', tone: 'success' });
      invalidateContent();
      await load();
      await loadVersions(active.id);
    } catch (saveError) {
      setError(toDisplayErrorMessage(saveError, 'Не удалось заменить файл'));
    } finally {
      setIsSaving(false);
    }
  };

  const remove = async (docId: string): Promise<void> => {
    if (!adminUserId) {
      return;
    }

    setIsSaving(true);
    setError(null);
    try {
      await deleteLegalDocRequest(adminUserId, docId);
      if (activeId === docId) {
        setActiveId('');
        setVersions([]);
      }
      showToast({ message: 'Документ удалён', tone: 'success' });
      invalidateContent();
      await load();
    } catch (saveError) {
      setError(toDisplayErrorMessage(saveError, 'Не удалось удалить документ'));
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <>
      <AdminPageHeader
        subtitle="Загрузка PDF и изображений. Идентификатор — латиница, цифры и дефис, например offer"
        title="Документы и соглашения"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Добавить документ</h2>
        <div className={clsx(styles.formGrid)}>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="legal-id">
              Идентификатор
            </label>
            <input
              className={clsx(styles.input)}
              id="legal-id"
              onChange={(event) =>
                setCreateDraft((previous) => ({ ...previous, id: event.target.value }))
              }
              placeholder="offer"
              value={createDraft.id}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="legal-name">
              Название
            </label>
            <input
              className={clsx(styles.input)}
              id="legal-name"
              onChange={(event) =>
                setCreateDraft((previous) => ({ ...previous, name: event.target.value }))
              }
              placeholder="Публичная оферта"
              value={createDraft.name}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="legal-create-version">
              Версия
            </label>
            <input
              className={clsx(styles.input)}
              id="legal-create-version"
              onChange={(event) =>
                setCreateDraft((previous) => ({ ...previous, version: event.target.value }))
              }
              value={createDraft.version}
            />
          </div>
          <div className={clsx(styles.field)}>
            <label className={clsx(styles.label)} htmlFor="legal-create-summary">
              Комментарий к версии
            </label>
            <input
              className={clsx(styles.input)}
              id="legal-create-summary"
              onChange={(event) =>
                setCreateDraft((previous) => ({ ...previous, summary: event.target.value }))
              }
              placeholder="Первая версия"
              value={createDraft.summary}
            />
          </div>
        </div>
        <DocumentFileField
          file={createDraft.file}
          id="legal-create-file"
          isDisabled={isSaving}
          isRequired
          label="Файл"
          onChange={(file) => setCreateDraft((previous) => ({ ...previous, file }))}
        />
        <div className={clsx(styles.actionsRow)}>
          <Button isDisabled={isSaving} onPress={() => void create()} variant="primary">
            {isSaving ? 'Сохраняем…' : 'Добавить'}
          </Button>
        </div>
      </section>

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Перечень</h2>
        {isLoading && docs.length === 0 ? (
          <p className={clsx(styles.hint)}>Загружаем документы…</p>
        ) : null}
        {!isLoading && docs.length === 0 ? (
          <p className={clsx(styles.hint)}>Документов пока нет</p>
        ) : null}
        {docs.length > 0 ? (
          <table className={clsx(styles.table)}>
            <thead>
              <tr>
                <th>Файл</th>
                <th>Название</th>
                <th>ID</th>
                <th>Версия</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {docs.map((doc) => (
                <tr key={doc.id}>
                  <td>
                    {doc.file_url ? (
                      <div className={clsx(styles.productThumbCell)}>
                        <FileThumb href={doc.file_url} />
                        <a href={doc.file_url} rel="noopener noreferrer" target="_blank">
                          Открыть
                        </a>
                      </div>
                    ) : (
                      '—'
                    )}
                  </td>
                  <td>
                    <button
                      className={clsx(
                        styles.smallButton,
                        doc.id === activeId && styles.smallButtonPrimary,
                      )}
                      onClick={() => setActiveId(doc.id)}
                      type="button"
                    >
                      {doc.name}
                    </button>
                  </td>
                  <td>{doc.id}</td>
                  <td>v{doc.current_version}</td>
                  <td>
                    <button
                      className={clsx(styles.smallButton, styles.smallButtonDanger)}
                      disabled={isSaving}
                      onClick={() => void remove(doc.id)}
                      type="button"
                    >
                      Удалить
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : null}
      </section>

      {active ? (
        <>
          <section className={clsx(styles.card)}>
            <h2 className={clsx(styles.cardTitle)}>Заменить файл: {active.name}</h2>
            <div className={clsx(styles.formGrid)}>
              <div className={clsx(styles.field)}>
                <label className={clsx(styles.label)} htmlFor="legal-replace-version">
                  Новая версия
                </label>
                <input
                  className={clsx(styles.input)}
                  id="legal-replace-version"
                  onChange={(event) =>
                    setReplaceDraft((previous) => ({ ...previous, version: event.target.value }))
                  }
                  placeholder={active.current_version}
                  value={replaceDraft.version}
                />
              </div>
              <div className={clsx(styles.field)}>
                <label className={clsx(styles.label)} htmlFor="legal-replace-summary">
                  Что изменилось
                </label>
                <input
                  className={clsx(styles.input)}
                  id="legal-replace-summary"
                  onChange={(event) =>
                    setReplaceDraft((previous) => ({ ...previous, summary: event.target.value }))
                  }
                  placeholder="Обновление документа"
                  value={replaceDraft.summary}
                />
              </div>
            </div>
            <DocumentFileField
              currentFileHref={active.file_url || undefined}
              currentFileLabel="Открыть текущий файл"
              file={replaceDraft.file}
              hint={`Предыдущие версии сохраняются в истории. ${DOCUMENT_FILE_HINT}`}
              id="legal-replace-file"
              isDisabled={isSaving}
              isRequired
              label="Новый файл"
              onChange={(file) => setReplaceDraft((previous) => ({ ...previous, file }))}
            />
            <div className={clsx(styles.actionsRow)}>
              <Button isDisabled={isSaving} onPress={() => void replace()} variant="primary">
                {isSaving ? 'Сохраняем…' : 'Заменить файл'}
              </Button>
            </div>
          </section>

          <section className={clsx(styles.card)}>
            <h2 className={clsx(styles.cardTitle)}>История версий</h2>
            {versions.length === 0 ? (
              <p className={clsx(styles.hint)}>История пока пуста</p>
            ) : (
              <table className={clsx(styles.table)}>
                <thead>
                  <tr>
                    <th>Файл</th>
                    <th>Версия</th>
                    <th>Дата</th>
                    <th>Изменения</th>
                    <th>Автор</th>
                  </tr>
                </thead>
                <tbody>
                  {versions.map((version) => (
                    <tr key={version.id}>
                      <td>
                        {version.file_url ? (
                          <div className={clsx(styles.productThumbCell)}>
                            <FileThumb href={version.file_url} />
                            <a href={version.file_url} rel="noopener noreferrer" target="_blank">
                              Скачать
                            </a>
                          </div>
                        ) : (
                          '—'
                        )}
                      </td>
                      <td>v{version.version}</td>
                      <td>{formatAdminDateTime(version.created_at)}</td>
                      <td>{version.summary}</td>
                      <td>{version.author || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </section>
        </>
      ) : null}
    </>
  );
};

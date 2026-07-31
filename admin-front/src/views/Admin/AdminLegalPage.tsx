'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect, useState } from 'react';

import { useContent } from '@/core/entities/content';

import styles from './Admin.module.css';
import { $contentSaveError, $isLegalSaving, legalSaveRequested } from './model/content';
import { AdminPageHeader } from './ui/AdminPageHeader';

type LegalDraft = {
  body: string;
  summary: string;
  version: string;
};

export const AdminLegalPage = (): JSX.Element => {
  const { legalDocs } = useContent();
  const [isSaving, error, save] = useUnit([
    $isLegalSaving,
    $contentSaveError,
    legalSaveRequested,
  ]);
  const [activeId, setActiveId] = useState('');
  const [draft, setDraft] = useState<LegalDraft>({ body: '', summary: '', version: '' });

  useEffect(() => {
    if (!activeId && legalDocs.length > 0) {
      setActiveId(legalDocs[0].id);
    }
  }, [activeId, legalDocs]);

  const active = legalDocs.find((doc) => doc.id === activeId);

  useEffect(() => {
    if (active) {
      setDraft({ body: active.body, summary: '', version: active.current_version });
    }
  }, [active]);

  return (
    <>
      <AdminPageHeader
        actions={
          active ? (
            <Button
              isDisabled={isSaving}
              onPress={() =>
                save({
                  body: draft.body,
                  docId: active.id,
                  summary: draft.summary || 'Обновление документа',
                  version: draft.version,
                })
              }
              variant="primary"
            >
              {isSaving ? 'Публикуем…' : 'Опубликовать'}
            </Button>
          ) : undefined
        }
        subtitle="Оферта, политика конфиденциальности, согласие на ПД и пользовательское соглашение"
        title="Юридические документы"
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <section className={clsx(styles.card)}>
        <div className={clsx(styles.actionsRow)}>
          {legalDocs.map((doc) => (
            <button
              className={clsx(styles.smallButton, doc.id === activeId && styles.smallButtonPrimary)}
              key={doc.id}
              onClick={() => setActiveId(doc.id)}
              type="button"
            >
              {doc.name}
            </button>
          ))}
        </div>
      </section>

      {active ? (
        <>
          <section className={clsx(styles.card)}>
            <div className={clsx(styles.formGrid)}>
              <div className={clsx(styles.field)}>
                <label className={clsx(styles.label)} htmlFor="legal-version">
                  Версия
                </label>
                <input
                  className={clsx(styles.input)}
                  id="legal-version"
                  onChange={(event) =>
                    setDraft((previous) => ({ ...previous, version: event.target.value }))
                  }
                  value={draft.version}
                />
              </div>
              <div className={clsx(styles.field)}>
                <label className={clsx(styles.label)} htmlFor="legal-summary">
                  Что изменилось
                </label>
                <input
                  className={clsx(styles.input)}
                  id="legal-summary"
                  onChange={(event) =>
                    setDraft((previous) => ({ ...previous, summary: event.target.value }))
                  }
                  placeholder="Например: уточнён порядок возврата"
                  value={draft.summary}
                />
              </div>
            </div>

            <div className={clsx(styles.field)}>
              <label className={clsx(styles.label)} htmlFor="legal-body">
                Текст документа
              </label>
              <textarea
                className={clsx(styles.textarea)}
                id="legal-body"
                onChange={(event) =>
                  setDraft((previous) => ({ ...previous, body: event.target.value }))
                }
                rows={18}
                value={draft.body}
              />
            </div>

            <p className={clsx(styles.hint)}>
              Если указать новый номер версии — создастся запись в истории версий. Старые версии не
              удаляются.
            </p>
          </section>

          <section className={clsx(styles.card)}>
            <h2 className={clsx(styles.cardTitle)}>История версий</h2>
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
                {active.versions.map((version) => (
                  <tr key={version.version}>
                    <td>v{version.version}</td>
                    <td>{version.date}</td>
                    <td>{version.summary}</td>
                    <td>{version.author}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </>
      ) : null}
    </>
  );
};

'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useState } from 'react';

import type {
  ContactsPageContent,
  ContentPageKey,
  HomePageContent,
  ListPageContent,
  TextPageContent,
} from '@/core/shared/api/content';

import { useContent } from '@/core/entities/content';
import { CONTENT_PAGE_TITLES } from '@/core/shared/api/content';

import styles from './Admin.module.css';
import { $contentSaveError, $isContentSaving, contentSaveRequested } from './model/content';
import { AdminPageHeader } from './ui/AdminPageHeader';
import { StringListEditor } from './ui/StringListEditor';

type AdminContentPageProps = {
  pageKey: ContentPageKey;
};

const isTextPage = (key: ContentPageKey): boolean => key === 'about' || key === 'terms';
const isListPage = (key: ContentPageKey): boolean => key === 'certificates' || key === 'promo';

export const AdminContentPage = ({ pageKey }: AdminContentPageProps): JSX.Element => {
  const { content } = useContent();
  const [isSaving, error, save] = useUnit([
    $isContentSaving,
    $contentSaveError,
    contentSaveRequested,
  ]);
  const [draft, setDraft] = useState<null | Record<string, unknown>>(null);
  const [draftKey, setDraftKey] = useState<ContentPageKey | null>(null);

  if (content && (draft === null || draftKey !== pageKey)) {
    setDraftKey(pageKey);
    setDraft({ ...content[pageKey] });
  }

  if (!draft) {
    return (
      <>
        <AdminPageHeader title={CONTENT_PAGE_TITLES[pageKey]} />
        <p className={clsx(styles.hint)}>Загружаем контент…</p>
      </>
    );
  }

  const patch = (value: Record<string, unknown>): void => {
    setDraft((previous) => ({ ...(previous ?? {}), ...value }));
  };

  const submit = (): void => {
    save({
      key: pageKey,
      value: draft as never,
    });
  };

  return (
    <>
      <AdminPageHeader
        actions={
          <Button isDisabled={isSaving} onPress={submit} variant="primary">
            {isSaving ? 'Сохраняем…' : 'Сохранить'}
          </Button>
        }
        subtitle="Изменения сразу попадают на публичную страницу портала"
        title={CONTENT_PAGE_TITLES[pageKey]}
      />

      {error ? <p className={clsx(styles.error)}>{error}</p> : null}

      <section className={clsx(styles.card)}>
        <div className={clsx(styles.form)}>
          {pageKey === 'home' ? (
            <HomeEditor onChange={patch} value={draft as unknown as HomePageContent} />
          ) : null}

          {isTextPage(pageKey) ? (
            <TextEditor onChange={patch} value={draft as unknown as TextPageContent} />
          ) : null}

          {isListPage(pageKey) ? (
            <ListEditor onChange={patch} value={draft as unknown as ListPageContent} />
          ) : null}

          {pageKey === 'contacts' ? (
            <ContactsEditor onChange={patch} value={draft as unknown as ContactsPageContent} />
          ) : null}
        </div>
      </section>
    </>
  );
};

type EditorProps<T> = {
  onChange: (value: Record<string, unknown>) => void;
  value: T;
};

const HomeEditor = ({ onChange, value }: EditorProps<HomePageContent>): JSX.Element => (
  <>
    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="home-hero-title">
        Заголовок баннера
      </label>
      <input
        className={clsx(styles.input)}
        id="home-hero-title"
        onChange={(event) => onChange({ hero_title: event.target.value })}
        value={value.hero_title}
      />
    </div>

    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="home-hero-subtitle">
        Подзаголовок
      </label>
      <textarea
        className={clsx(styles.textarea)}
        id="home-hero-subtitle"
        onChange={(event) => onChange({ hero_subtitle: event.target.value })}
        value={value.hero_subtitle}
      />
    </div>

    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="home-hero-button">
        Текст кнопки
      </label>
      <input
        className={clsx(styles.input)}
        id="home-hero-button"
        onChange={(event) => onChange({ hero_button: event.target.value })}
        value={value.hero_button}
      />
    </div>

    <StringListEditor
      label="Преимущества"
      onChange={(features) => onChange({ features })}
      placeholder="Например: более 4 200 позиций"
      values={value.features}
    />

    <div className={clsx(styles.field)}>
      <span className={clsx(styles.label)}>Цифры о компании</span>
      <div className={clsx(styles.listEditor)}>
        {value.stats.map((stat, index) => (
          <div className={clsx(styles.listRow)} key={`stat-${index}`}>
            <input
              aria-label={`Значение ${index + 1}`}
              className={clsx(styles.input)}
              onChange={(event) =>
                onChange({
                  stats: value.stats.map((item, itemIndex) =>
                    itemIndex === index ? { ...item, value: event.target.value } : item,
                  ),
                })
              }
              placeholder="4 200+"
              value={stat.value}
            />
            <input
              aria-label={`Подпись ${index + 1}`}
              className={clsx(styles.input)}
              onChange={(event) =>
                onChange({
                  stats: value.stats.map((item, itemIndex) =>
                    itemIndex === index ? { ...item, label: event.target.value } : item,
                  ),
                })
              }
              placeholder="SKU в каталоге"
              value={stat.label}
            />
            <button
              className={clsx(styles.smallButton, styles.smallButtonDanger)}
              onClick={() =>
                onChange({ stats: value.stats.filter((_, itemIndex) => itemIndex !== index) })
              }
              type="button"
            >
              Удалить
            </button>
          </div>
        ))}
        <button
          className={clsx(styles.smallButton)}
          onClick={() => onChange({ stats: [...value.stats, { label: '', value: '' }] })}
          type="button"
        >
          Добавить показатель
        </button>
      </div>
    </div>

    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="home-why-title">
        Заголовок блока «Почему выбирают нас»
      </label>
      <input
        className={clsx(styles.input)}
        id="home-why-title"
        onChange={(event) => onChange({ why_title: event.target.value })}
        value={value.why_title}
      />
    </div>

    <StringListEditor
      label="Пункты блока «Почему выбирают нас»"
      onChange={(whyItems) => onChange({ why_items: whyItems })}
      values={value.why_items}
    />
  </>
);

const TextEditor = ({ onChange, value }: EditorProps<TextPageContent>): JSX.Element => (
  <>
    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="text-title">
        Заголовок страницы
      </label>
      <input
        className={clsx(styles.input)}
        id="text-title"
        onChange={(event) => onChange({ title: event.target.value })}
        value={value.title}
      />
    </div>
    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="text-body">
        Текст (пустая строка — новый абзац)
      </label>
      <textarea
        className={clsx(styles.textarea)}
        id="text-body"
        onChange={(event) => onChange({ text: event.target.value })}
        rows={14}
        value={value.text}
      />
    </div>
  </>
);

const ListEditor = ({ onChange, value }: EditorProps<ListPageContent>): JSX.Element => (
  <>
    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="list-title">
        Заголовок страницы
      </label>
      <input
        className={clsx(styles.input)}
        id="list-title"
        onChange={(event) => onChange({ title: event.target.value })}
        value={value.title}
      />
    </div>
    <div className={clsx(styles.field)}>
      <label className={clsx(styles.label)} htmlFor="list-text">
        Вводный текст
      </label>
      <textarea
        className={clsx(styles.textarea)}
        id="list-text"
        onChange={(event) => onChange({ text: event.target.value })}
        value={value.text}
      />
    </div>
    <StringListEditor
      label="Пункты списка"
      onChange={(items) => onChange({ items })}
      values={value.items}
    />
  </>
);

const ContactsEditor = ({ onChange, value }: EditorProps<ContactsPageContent>): JSX.Element => {
  const offices = value.offices ?? [];
  const requisiteItems = value.requisite_items ?? [];

  return (
    <>
      <div className={clsx(styles.formGrid)}>
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="contacts-title">
            Заголовок
          </label>
          <input
            className={clsx(styles.input)}
            id="contacts-title"
            onChange={(event) => onChange({ title: event.target.value })}
            value={value.title}
          />
        </div>
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="contacts-phone">
            Телефон
          </label>
          <input
            className={clsx(styles.input)}
            id="contacts-phone"
            onChange={(event) => onChange({ phone: event.target.value })}
            value={value.phone}
          />
        </div>
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="contacts-email">
            E-mail
          </label>
          <input
            className={clsx(styles.input)}
            id="contacts-email"
            onChange={(event) => onChange({ email: event.target.value })}
            value={value.email}
          />
        </div>
        <div className={clsx(styles.field)}>
          <label className={clsx(styles.label)} htmlFor="contacts-hours">
            Часы работы
          </label>
          <input
            className={clsx(styles.input)}
            id="contacts-hours"
            onChange={(event) => onChange({ work_hours: event.target.value })}
            value={value.work_hours}
          />
        </div>
      </div>

      <div className={clsx(styles.field)}>
        <label className={clsx(styles.label)} htmlFor="contacts-subtitle">
          Подзаголовок
        </label>
        <input
          className={clsx(styles.input)}
          id="contacts-subtitle"
          onChange={(event) => onChange({ subtitle: event.target.value })}
          value={value.subtitle ?? ''}
        />
      </div>

      <div className={clsx(styles.field)}>
        <label className={clsx(styles.label)} htmlFor="contacts-address">
          Адрес
        </label>
        <input
          className={clsx(styles.input)}
          id="contacts-address"
          onChange={(event) => onChange({ address: event.target.value })}
          value={value.address}
        />
      </div>

      <div className={clsx(styles.field)}>
        <label className={clsx(styles.label)} htmlFor="contacts-requisites">
          Реквизиты (кратко)
        </label>
        <textarea
          className={clsx(styles.textarea)}
          id="contacts-requisites"
          onChange={(event) => onChange({ requisites: event.target.value })}
          value={value.requisites}
        />
      </div>

      <div className={clsx(styles.field)}>
        <span className={clsx(styles.label)}>Отделы</span>
        <div className={clsx(styles.listEditor)}>
          {value.managers.map((manager, index) => (
            <div className={clsx(styles.formGrid)} key={`manager-${index}`}>
              <input
                aria-label={`Название отдела ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    managers: value.managers.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, name: event.target.value } : item,
                    ),
                  })
                }
                placeholder="Отдел продаж"
                value={manager.name}
              />
              <input
                aria-label={`Описание ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    managers: value.managers.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, role: event.target.value } : item,
                    ),
                  })
                }
                placeholder="Оформление заказов"
                value={manager.role}
              />
              <input
                aria-label={`Телефон ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    managers: value.managers.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, phone: event.target.value } : item,
                    ),
                  })
                }
                placeholder="+7 846 000-00-00"
                value={manager.phone}
              />
              <input
                aria-label={`E-mail ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    managers: value.managers.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, email: event.target.value } : item,
                    ),
                  })
                }
                placeholder="order@voint.ru"
                value={manager.email}
              />
              <button
                className={clsx(styles.smallButton, styles.smallButtonDanger)}
                onClick={() =>
                  onChange({
                    managers: value.managers.filter((_, itemIndex) => itemIndex !== index),
                  })
                }
                type="button"
              >
                Удалить
              </button>
            </div>
          ))}
          <button
            className={clsx(styles.smallButton)}
            onClick={() =>
              onChange({
                managers: [...value.managers, { email: '', name: '', phone: '', role: '' }],
              })
            }
            type="button"
          >
            Добавить отдел
          </button>
        </div>
      </div>

      <div className={clsx(styles.field)}>
        <span className={clsx(styles.label)}>Представительства</span>
        <div className={clsx(styles.listEditor)}>
          {offices.map((office, index) => (
            <div className={clsx(styles.formGrid)} key={`office-${index}`}>
              <input
                aria-label={`Город ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    offices: offices.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, city: event.target.value } : item,
                    ),
                  })
                }
                placeholder="Самара"
                value={office.city}
              />
              <input
                aria-label={`Адрес ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    offices: offices.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, address: event.target.value } : item,
                    ),
                  })
                }
                placeholder="Адрес"
                value={office.address}
              />
              <input
                aria-label={`Телефон офиса ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    offices: offices.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, phone: event.target.value } : item,
                    ),
                  })
                }
                placeholder="+7 846 000-00-00"
                value={office.phone}
              />
              <input
                aria-label={`E-mail офиса ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    offices: offices.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, email: event.target.value } : item,
                    ),
                  })
                }
                placeholder="office@voint.ru"
                value={office.email}
              />
              <label className={clsx(styles.checkboxLabel)}>
                <input
                  checked={Boolean(office.is_main)}
                  onChange={(event) =>
                    onChange({
                      offices: offices.map((item, itemIndex) =>
                        itemIndex === index ? { ...item, is_main: event.target.checked } : item,
                      ),
                    })
                  }
                  type="checkbox"
                />
                Главный офис
              </label>
              <button
                className={clsx(styles.smallButton, styles.smallButtonDanger)}
                onClick={() =>
                  onChange({
                    offices: offices.filter((_, itemIndex) => itemIndex !== index),
                  })
                }
                type="button"
              >
                Удалить
              </button>
            </div>
          ))}
          <button
            className={clsx(styles.smallButton)}
            onClick={() =>
              onChange({
                offices: [...offices, { address: '', city: '', email: '', phone: '' }],
              })
            }
            type="button"
          >
            Добавить представительство
          </button>
        </div>
      </div>

      <div className={clsx(styles.field)}>
        <span className={clsx(styles.label)}>Реквизиты (строки)</span>
        <div className={clsx(styles.listEditor)}>
          {requisiteItems.map((row, index) => (
            <div className={clsx(styles.formGrid)} key={`requisite-${index}`}>
              <input
                aria-label={`Поле реквизита ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    requisite_items: requisiteItems.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, label: event.target.value } : item,
                    ),
                  })
                }
                placeholder="ИНН"
                value={row.label}
              />
              <input
                aria-label={`Значение реквизита ${index + 1}`}
                className={clsx(styles.input)}
                onChange={(event) =>
                  onChange({
                    requisite_items: requisiteItems.map((item, itemIndex) =>
                      itemIndex === index ? { ...item, value: event.target.value } : item,
                    ),
                  })
                }
                placeholder="0000000000"
                value={row.value}
              />
              <button
                className={clsx(styles.smallButton, styles.smallButtonDanger)}
                onClick={() =>
                  onChange({
                    requisite_items: requisiteItems.filter((_, itemIndex) => itemIndex !== index),
                  })
                }
                type="button"
              >
                Удалить
              </button>
            </div>
          ))}
          <button
            className={clsx(styles.smallButton)}
            onClick={() =>
              onChange({
                requisite_items: [...requisiteItems, { label: '', value: '' }],
              })
            }
            type="button"
          >
            Добавить строку
          </button>
        </div>
      </div>
    </>
  );
};

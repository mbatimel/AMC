'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useState } from 'react';

import type {
  AboutPageContent,
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
import { HtmlEditor } from './ui/HtmlEditor';
import { StringListEditor } from './ui/StringListEditor';

type AdminContentPageProps = {
  pageKey: ContentPageKey;
};

const isTextPage = (key: ContentPageKey): boolean => key === 'terms';
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

          {pageKey === 'about' ? (
            <AboutEditor onChange={patch} value={draft as unknown as AboutPageContent} />
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

const field = (
  id: string,
  label: string,
  value: string,
  onChange: (next: string) => void,
): JSX.Element => (
  <div className={clsx(styles.field)}>
    <label className={clsx(styles.label)} htmlFor={id}>
      {label}
    </label>
    <input
      className={clsx(styles.input)}
      id={id}
      onChange={(event) => onChange(event.target.value)}
      value={value}
    />
  </div>
);

const AboutEditor = ({ onChange, value }: EditorProps<AboutPageContent>): JSX.Element => (
  <>
    {field('about-title', 'Заголовок страницы', value.title, (title) => onChange({ title }))}
    {field('about-hero-badge', 'Бейдж в шапке', value.hero_badge ?? '', (hero_badge) =>
      onChange({ hero_badge }),
    )}
    {field(
      'about-hero-subtitle',
      'Подзаголовок в шапке',
      value.hero_subtitle ?? '',
      (hero_subtitle) => onChange({ hero_subtitle }),
    )}
    {field(
      'about-profile-badge',
      'Бейдж блока «Кто мы»',
      value.profile_badge ?? '',
      (profile_badge) => onChange({ profile_badge }),
    )}
    {field(
      'about-profile-title',
      'Заголовок блока «Кто мы»',
      value.profile_title ?? '',
      (profile_title) => onChange({ profile_title }),
    )}
    <HtmlEditor
      label="Текст «Кто мы»"
      onChange={(text) => onChange({ text })}
      rows={10}
      value={value.text}
    />
    {field(
      'about-directions-badge',
      'Бейдж «Ассортимент»',
      value.directions_badge ?? '',
      (directions_badge) => onChange({ directions_badge }),
    )}
    {field(
      'about-directions-title',
      'Заголовок «Ассортимент»',
      value.directions_title ?? '',
      (directions_title) => onChange({ directions_title }),
    )}
    {field(
      'about-directions-subtitle',
      'Подзаголовок «Ассортимент»',
      value.directions_subtitle ?? '',
      (directions_subtitle) => onChange({ directions_subtitle }),
    )}
    {field('about-offices-badge', 'Бейдж «География»', value.offices_badge ?? '', (offices_badge) =>
      onChange({ offices_badge }),
    )}
    {field(
      'about-offices-title',
      'Заголовок «География»',
      value.offices_title ?? '',
      (offices_title) => onChange({ offices_title }),
    )}
    {field(
      'about-offices-subtitle',
      'Подзаголовок «География»',
      value.offices_subtitle ?? '',
      (offices_subtitle) => onChange({ offices_subtitle }),
    )}
    {field('about-cta-badge', 'Бейдж призыва', value.cta_badge ?? '', (cta_badge) =>
      onChange({ cta_badge }),
    )}
    {field('about-cta-title', 'Заголовок призыва', value.cta_title ?? '', (cta_title) =>
      onChange({ cta_title }),
    )}
    {field('about-cta-text', 'Текст призыва', value.cta_text ?? '', (cta_text) =>
      onChange({ cta_text }),
    )}
    {field('about-cta-button', 'Текст кнопки заявки', value.cta_button ?? '', (cta_button) =>
      onChange({ cta_button }),
    )}
    {field('about-cta-hint', 'Подсказка про email', value.cta_hint ?? '', (cta_hint) =>
      onChange({ cta_hint }),
    )}
  </>
);

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

    <HtmlEditor
      label="Подзаголовок"
      onChange={(hero_subtitle) => onChange({ hero_subtitle })}
      rows={4}
      value={value.hero_subtitle}
    />

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
    <HtmlEditor
      label="Текст"
      onChange={(text) => onChange({ text })}
      rows={14}
      value={value.text}
    />
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
    <HtmlEditor label="Вводный текст" onChange={(text) => onChange({ text })} value={value.text} />
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
        <label className={clsx(styles.label)} htmlFor="contacts-warehouse-caption">
          Адрес склада / подпись к карте
        </label>
        <input
          className={clsx(styles.input)}
          id="contacts-warehouse-caption"
          onChange={(event) => onChange({ warehouse_map_caption: event.target.value })}
          value={value.warehouse_map_caption ?? ''}
        />
      </div>

      <div className={clsx(styles.field)}>
        <label className={clsx(styles.label)} htmlFor="contacts-warehouse-map">
          Схема проезда на склад (URL виджета Яндекс.Карт)
        </label>
        <input
          className={clsx(styles.input)}
          id="contacts-warehouse-map"
          onChange={(event) => onChange({ warehouse_map_url: event.target.value })}
          placeholder="https://yandex.ru/map-widget/v1/?ll=..."
          value={value.warehouse_map_url ?? ''}
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

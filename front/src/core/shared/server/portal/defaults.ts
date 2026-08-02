import type { PortalState, Promotion } from './types';

const nowIso = (): string => new Date().toISOString();

const pad = (value: number): string => String(value).padStart(2, '0');

/** Формат для `<input type="datetime-local">`: YYYY-MM-DDTHH:mm */
const toDateTimeLocal = (date: Date): string =>
  `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;

const daysFromNow = (days: number): Date => {
  const date = new Date();

  date.setDate(date.getDate() + days);

  return date;
};

const seedPromotions = (): Promotion[] => [
  {
    condition: 'Скидка действует на выбранные товары в течение периода',
    createdAt: nowIso(),
    desc: 'Сниженная цена на все метчики в течение двух недель.',
    discMode: 'percent',
    discValue: 15,
    endAt: toDateTimeLocal(daysFromNow(14)),
    endedManually: false,
    id: 'promo-metchiki',
    minQty: 0,
    name: 'Скидка на метчики',
    sel: { all: false, nodes: [], products: [] },
    startAt: toDateTimeLocal(daysFromNow(-1)),
    type: 'date',
  },
  {
    condition: 'Скидка 10% при покупке от 3 шт выбранного товара',
    createdAt: nowIso(),
    desc: 'Скидка при заказе от 3 единиц одного наименования.',
    discMode: 'percent',
    discValue: 10,
    endAt: toDateTimeLocal(daysFromNow(30)),
    endedManually: false,
    id: 'promo-sverla-qty',
    minQty: 3,
    name: 'Свёрла оптом — выгоднее',
    sel: { all: false, nodes: [], products: [] },
    startAt: toDateTimeLocal(daysFromNow(-2)),
    type: 'qty',
  },
];

export const createDefaultPortalState = (): PortalState => ({
  audit_log: [
    {
      action: 'Портал запущен, загружены начальные данные контента',
      actor_label: 'Система',
      created_at: nowIso(),
      id: 'audit-seed',
    },
  ],
  banners: {
    delay_sec: 6,
    items: [
      {
        dateFrom: '',
        dateTo: '',
        id: 'banner-drills',
        image: '',
        is_active: true,
        link: '/catalog?q=%D1%81%D0%B2%D0%B5%D1%80%D0%BB%D0%BE',
        sort_order: 1,
        subtitle: 'Промышленная серия для точной обработки',
        title: 'Свёрла ВИ-Про',
      },
      {
        dateFrom: '',
        dateTo: '',
        id: 'banner-taps',
        image: '',
        is_active: true,
        link: '/catalog?gost=%D0%93%D0%9E%D0%A1%D0%A2%203266-81',
        sort_order: 2,
        subtitle: 'Подбор по ГОСТ, размеру и материалу',
        title: 'Метчики Р6М5',
      },
      {
        dateFrom: '',
        dateTo: '',
        id: 'banner-delivery',
        image: '',
        is_active: true,
        link: '/terms',
        sort_order: 3,
        subtitle: 'Отгрузка со склада в день оплаты',
        title: 'Доставка по России',
      },
      {
        dateFrom: '',
        dateTo: '',
        id: 'banner-dies',
        image: '',
        is_active: true,
        link: '/catalog?q=%D0%BF%D0%BB%D0%B0%D1%88%D0%BA%D0%B0',
        sort_order: 4,
        subtitle: 'Стабильная геометрия резьбы',
        title: 'Плашки 9ХС',
      },
    ],
  },
  content: {
    about: {
      text: [
        'ООО ПО «Волжский инструмент» — производитель профессионального режущего инструмента с 1998 года. Предприятие расположено в г. Самара и выпускает свёрла, фрезы, метчики, плашки и специальный инструмент по ГОСТ и ТУ.',
        'Основные направления: серийное производство стандартного инструмента и изготовление специального инструмента по чертежам заказчика.',
        'Продукция поставляется на предприятия машиностроения, авиастроения, нефтегазовой отрасли и оборонного комплекса.',
      ].join('\n\n'),
      title: 'О компании',
    },
    certificates: {
      items: [
        'ГОСТ 10902-77 — свёрла спиральные с цилиндрическим хвостовиком',
        'ГОСТ 3266-81 — метчики машинно-ручные',
        'ГОСТ 9740-71 — плашки круглые',
        'ГОСТ 17025-71 — развёртки',
        'ТУ 3918-001 — специальный инструмент по чертежам заказчика',
        'ISO 9001:2015 — система менеджмента качества',
      ],
      text: 'Вся продукция сертифицирована по ГОСТ и ТУ. Копии сертификатов и паспортов качества предоставляются по запросу и прикладываются к отгрузочным документам.',
      title: 'Сертификаты и лицензии',
    },
    contacts: {
      address: '443004, г. Самара, ул. Промышленная, д. 7',
      email: 'order@amc.ru',
      managers: [
        {
          email: 'order@amc.ru',
          name: 'Отдел продаж',
          phone: '+7 846 265-93-10',
          role: 'Оформление и сопровождение заказов',
        },
        {
          email: 'support@amc.ru',
          name: 'Техническая поддержка',
          phone: '+7 846 265-93-10 доб. 331-37-15',
          role: 'Подбор инструмента, вопросы по порталу',
        },
      ],
      phone: '+7 846 265-93-10',
      requisites: 'ООО ПО «Волжский инструмент», ИНН 6316000000, КПП 631601001, ОГРН 1026300000000',
      title: 'Контакты',
      work_hours: 'Пн–Пт: 08:00–17:00 (МСК+1)',
    },
    home: {
      features: [
        'Более 4 200 позиций — собственное производство и партнёрские бренды',
        'Индивидуальные цены и отсрочка для оптовых клиентов',
        'Отгрузка со склада и доставка по всей России',
      ],
      hero_button: 'Перейти в каталог',
      hero_subtitle:
        'Метчики, свёрла, плашки и слесарный инструмент собственного производства. Реальные остатки, индивидуальные цены и закрывающие документы — всё онлайн.',
      hero_title: 'Профессиональный инструмент оптом',
      stats: [
        { label: 'SKU в каталоге', value: '4 200+' },
        { label: 'оптовых клиентов', value: '820+' },
        { label: 'года на рынке', value: '33' },
        { label: 'заказов в срок', value: '99,2%' },
      ],
      why_items: [
        'ГОСТ и ТУ на каждое изделие',
        'Индивидуальные условия для оптовиков',
        'Техническая поддержка и подбор инструмента',
      ],
      why_title: 'Почему выбирают нас',
    },
    promo: {
      items: [
        'Скидка 7% на кобальтовые свёрла при заказе от 50 000 ₽',
        'Комплекты метчиков М6–М12 по цене оптовой группы для новых клиентов',
        'Бесплатная доставка ТК при заказе от 100 000 ₽',
      ],
      text: 'Действующие акции и специальные условия для оптовых клиентов портала. Условия суммируются с индивидуальной скидкой прайс-группы.',
      title: 'Акции и специальные предложения',
    },
    terms: {
      text: [
        'Минимальный заказ: от 5 000 ₽ (без НДС).',
        'Оплата: безналичный расчёт, предоплата или отсрочка по договору.',
        'Доставка: транспортная компания, самовывоз со склада в Самаре.',
        'Возврат: в течение 14 дней при сохранении товарного вида и упаковки.',
        'Индивидуальные цены назначаются после подтверждения заявки менеджером.',
      ].join('\n'),
      title: 'Условия работы',
    },
  },
  feedback: [],
  legal_docs: [
    {
      body: 'Настоящий документ является публичной офертой ООО ПО «Волжский инструмент» о заключении договора поставки на условиях, размещённых на портале.',
      current_version: '1.0',
      id: 'offer',
      name: 'Договор оферты',
      updated_at: nowIso(),
      versions: [
        {
          author: 'Админ портала',
          date: '15.01.2026',
          summary: 'Первая версия оферты',
          version: '1.0',
        },
      ],
    },
    {
      body: 'Политика описывает состав обрабатываемых персональных данных, цели обработки, сроки хранения и права субъекта персональных данных.',
      current_version: '1.0',
      id: 'privacy',
      name: 'Политика конфиденциальности',
      updated_at: nowIso(),
      versions: [
        { author: 'Админ портала', date: '15.01.2026', summary: 'Первая версия', version: '1.0' },
      ],
    },
    {
      body: 'Согласие на обработку персональных данных предоставляется пользователем при регистрации на портале.',
      current_version: '1.0',
      id: 'consent',
      name: 'Согласие на обработку ПД',
      updated_at: nowIso(),
      versions: [
        { author: 'Админ портала', date: '15.01.2026', summary: 'Первая версия', version: '1.0' },
      ],
    },
    {
      body: 'Пользовательское соглашение регулирует порядок использования портала, личного кабинета и электронных заказов.',
      current_version: '1.0',
      id: 'user-agreement',
      name: 'Пользовательское соглашение',
      updated_at: nowIso(),
      versions: [
        { author: 'Админ портала', date: '15.01.2026', summary: 'Первая версия', version: '1.0' },
      ],
    },
  ],
  portal_users: [],
  promotions: seedPromotions(),
  signup_requests: [],
  support_requests: [],
});

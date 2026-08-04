import type { Icon } from '@/core/shared/icons';

export type AboutDirection = {
  description: string;
  Icon: Icon;
  title: string;
};

export type AboutOffice = {
  city: string;
  description: string;
  isMain?: boolean;
};

export const ABOUT_HERO_BADGE = 'Производство и поставка инструмента';

export const ABOUT_HERO_SUBTITLE =
  '33 года на рынке металлорежущего инструмента. Собственное производство в Самаре и шесть представительств по России.';

export const ABOUT_PROFILE_BADGE = 'Профиль';
export const ABOUT_PROFILE_TITLE = 'Кто мы и что делаем';

export const ABOUT_DIRECTIONS_BADGE = 'Ассортимент';
export const ABOUT_DIRECTIONS_TITLE = 'Шесть направлений работы';
export const ABOUT_DIRECTIONS_SUBTITLE = 'Полный спектр инструмента для промышленности и стройки';

export const ABOUT_OFFICES_BADGE = 'География';
export const ABOUT_OFFICES_TITLE = 'Шесть представительств по России';
export const ABOUT_OFFICES_SUBTITLE = 'Самовывоз доступен в любом городе';

export const ABOUT_CTA_BADGE = 'Сотрудничество';
export const ABOUT_CTA_TITLE = 'Готовы начать работу?';
export const ABOUT_CTA_TEXT =
  'Подберём и просчитаем инструмент под ваш производственный процесс или на полку в магазин.';

const IconMetal = ({
  className,
  currentColor = 'currentColor',
  height = 20,
  width = 20,
}: {
  className?: string;
  currentColor?: string;
  height?: number;
  width?: number;
}): JSX.Element => (
  <svg
    aria-hidden
    className={className}
    fill="none"
    height={height}
    viewBox="0 0 20 20"
    width={width}
  >
    <path
      d="M4 14.5 10 3.5l6 11H4Z"
      stroke={currentColor}
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
    <path d="M7.5 14.5h5" stroke={currentColor} strokeLinecap="round" strokeWidth="1.5" />
  </svg>
);

const IconWrench = ({
  className,
  currentColor = 'currentColor',
  height = 20,
  width = 20,
}: {
  className?: string;
  currentColor?: string;
  height?: number;
  width?: number;
}): JSX.Element => (
  <svg
    aria-hidden
    className={className}
    fill="none"
    height={height}
    viewBox="0 0 20 20"
    width={width}
  >
    <path
      d="M12.5 3.5a3.5 3.5 0 0 0-4.7 4.7L3.5 12.5v4h4l4.3-4.3a3.5 3.5 0 0 0 4.7-4.7l-2.5 2.5-2.5-2.5 2.5-2.5Z"
      stroke={currentColor}
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
  </svg>
);

const IconGarden = ({
  className,
  currentColor = 'currentColor',
  height = 20,
  width = 20,
}: {
  className?: string;
  currentColor?: string;
  height?: number;
  width?: number;
}): JSX.Element => (
  <svg
    aria-hidden
    className={className}
    fill="none"
    height={height}
    viewBox="0 0 20 20"
    width={width}
  >
    <path
      d="M10 17V9M10 9c0-3 2-5.5 5-6-1 3-3 5-5 6ZM10 9c0-3-2-5.5-5-6 1 3 3 5 5 6Z"
      stroke={currentColor}
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
  </svg>
);

const IconConsumables = ({
  className,
  currentColor = 'currentColor',
  height = 20,
  width = 20,
}: {
  className?: string;
  currentColor?: string;
  height?: number;
  width?: number;
}): JSX.Element => (
  <svg
    aria-hidden
    className={className}
    fill="none"
    height={height}
    viewBox="0 0 20 20"
    width={width}
  >
    <circle cx="10" cy="10" r="6.5" stroke={currentColor} strokeWidth="1.5" />
    <path d="M10 6.5v7M6.5 10h7" stroke={currentColor} strokeLinecap="round" strokeWidth="1.5" />
  </svg>
);

const IconBuild = ({
  className,
  currentColor = 'currentColor',
  height = 20,
  width = 20,
}: {
  className?: string;
  currentColor?: string;
  height?: number;
  width?: number;
}): JSX.Element => (
  <svg
    aria-hidden
    className={className}
    fill="none"
    height={height}
    viewBox="0 0 20 20"
    width={width}
  >
    <path
      d="M4 16V8l6-4 6 4v8M7.5 16v-4h5v4"
      stroke={currentColor}
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
  </svg>
);

const IconShield = ({
  className,
  currentColor = 'currentColor',
  height = 20,
  width = 20,
}: {
  className?: string;
  currentColor?: string;
  height?: number;
  width?: number;
}): JSX.Element => (
  <svg
    aria-hidden
    className={className}
    fill="none"
    height={height}
    viewBox="0 0 20 20"
    width={width}
  >
    <path
      d="M10 3.5 15.5 5.5v4.2c0 3.3-2.2 5.7-5.5 6.8-3.3-1.1-5.5-3.5-5.5-6.8V5.5L10 3.5Z"
      stroke={currentColor}
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
  </svg>
);

export const ABOUT_DIRECTIONS: AboutDirection[] = [
  {
    description: 'Метчики, плашки, свёрла, фрезы из Р6М5 и Р18К5',
    Icon: IconMetal,
    title: 'Металлорежущий',
  },
  {
    description: 'Наборы ключей, отвёртки, плоскогубцы и другой инструмент',
    Icon: IconWrench,
    title: 'Слесарный',
  },
  {
    description: 'Лопаты, секаторы, грабли, инвентарь для сада',
    Icon: IconGarden,
    title: 'Садовый',
  },
  {
    description: 'Диски отрезные, пилки, биты, расходный инструмент',
    Icon: IconConsumables,
    title: 'Расходники',
  },
  {
    description: 'Молотки, рулетки, уровни, строительный инструмент',
    Icon: IconBuild,
    title: 'Строительный',
  },
  {
    description: 'Перчатки, спецодежда, средства индивидуальной защиты',
    Icon: IconShield,
    title: 'Защита',
  },
];

export const ABOUT_OFFICES: AboutOffice[] = [
  {
    city: 'Самара',
    description: 'Головной офис, производство, основной склад. ул. Промышленная, д. 7',
    isMain: true,
  },
  {
    city: 'Москва',
    description: 'Логистический центр. Можайское шоссе, д. 2',
  },
  {
    city: 'Санкт-Петербург',
    description: 'Северо-Западный филиал. пр. Обуховской Обороны, д. 11',
  },
  {
    city: 'Екатеринбург',
    description: 'Уральский филиал. ул. Фронтовых бригад, д. 18',
  },
  {
    city: 'Челябинск',
    description: 'Складская площадка. ул. Троицкий тракт, д. 11',
  },
  {
    city: 'Казань',
    description: 'Поволжское представительство. ул. Магистральная, д. 23',
  },
];

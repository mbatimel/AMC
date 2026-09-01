export const toTelHref = (phone: string): string => `tel:${phone.replace(/[^\d+]/g, '')}`;

const YANDEX_MAP_HOSTS = new Set(['maps.yandex.ru', 'yandex.com', 'yandex.ru']);

export const isYandexMapEmbedUrl = (value: string): boolean => {
  try {
    const url = new URL(value);

    return url.protocol === 'https:' && YANDEX_MAP_HOSTS.has(url.hostname);
  } catch {
    return false;
  }
};

export const formatRequisitesText = (items: { label: string; value: string }[]): string =>
  items.map((row) => `${row.label}: ${row.value}`).join('\n');

export const downloadRequisitesFile = (text: string, filename = 'requisites.txt'): void => {
  const blob = new Blob([`\uFEFF${text}\n`], { type: 'text/plain;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');

  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
};

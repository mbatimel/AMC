const ALLOWED_TAGS = new Set(['a', 'b', 'br', 'em', 'i', 'li', 'ol', 'p', 'span', 'strong', 'ul']);

const ALLOWED_SIZE_CLASSES = new Set(['rich-text-size-l', 'rich-text-size-m', 'rich-text-size-s']);

const isSafeHref = (value: string): boolean =>
  /^(https?:|mailto:|\/|#)/i.test(value.trim()) && !/^\s*javascript:/i.test(value);

export const escapeText = (value: string): string =>
  value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;');

const readAttr = (rawAttrs: string, name: string): string => {
  const match = rawAttrs.match(
    new RegExp(`\\b${name}\\s*=\\s*(?:"([^"]*)"|'([^']*)'|([^\\s>]+))`, 'i'),
  );

  return match?.[1] ?? match?.[2] ?? match?.[3] ?? '';
};

const sanitizeAttributes = (tag: string, rawAttrs: string): string => {
  if (tag === 'a') {
    const href = readAttr(rawAttrs, 'href');

    if (!href || !isSafeHref(href)) {
      return '';
    }

    return ` href="${escapeText(href)}"`;
  }

  if (tag === 'span') {
    const className = readAttr(rawAttrs, 'class');
    const sizeClass = className
      .split(/\s+/)
      .map((item) => item.trim())
      .find((item) => ALLOWED_SIZE_CLASSES.has(item));

    return sizeClass ? ` class="${sizeClass}"` : '';
  }

  return '';
};

/** Оставляет только безопасные теги: жирный, курсив, ссылки, списки, абзацы, размер. */
export const sanitizeHtml = (dirty: string): string => {
  const withoutDanger = dirty
    .replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
    .replace(/<style[\s\S]*?>[\s\S]*?<\/style>/gi, '')
    .replace(/\son\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '');

  return withoutDanger.replace(
    /<\/?([a-z][a-z0-9]*)\b([^>]*)\/?>/gi,
    (match, tagName: string, attrs: string) => {
      const tag = tagName.toLowerCase();

      if (!ALLOWED_TAGS.has(tag)) {
        return '';
      }

      if (match.startsWith('</')) {
        return `</${tag}>`;
      }

      if (tag === 'br') {
        return '<br />';
      }

      if (tag === 'span') {
        const safeAttrs = sanitizeAttributes(tag, attrs);

        return safeAttrs ? `<span${safeAttrs}>` : '';
      }

      return `<${tag}${sanitizeAttributes(tag, attrs)}>`;
    },
  );
};

export const looksLikeHtml = (value: string): boolean => /<[a-z][\s\S]*>/i.test(value);

/** Текст/HTML из стора → HTML для TipTap. */
export const toRichTextHtml = (value: string): string => {
  if (!value.trim()) {
    return '';
  }

  if (looksLikeHtml(value)) {
    return sanitizeHtml(value);
  }

  return value
    .split(/\n+/)
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => `<p>${escapeText(line)}</p>`)
    .join('');
};

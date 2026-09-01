const ALLOWED_TAGS = new Set(['a', 'b', 'br', 'em', 'i', 'li', 'ol', 'p', 'strong', 'ul']);

const isSafeHref = (value: string): boolean =>
  /^(https?:|mailto:|\/|#)/i.test(value.trim()) && !/^\s*javascript:/i.test(value);

const escapeText = (value: string): string =>
  value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;');

const sanitizeAttributes = (tag: string, rawAttrs: string): string => {
  if (tag !== 'a') {
    return '';
  }

  const hrefMatch = rawAttrs.match(/\bhref\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))/i);
  const href = hrefMatch?.[1] ?? hrefMatch?.[2] ?? hrefMatch?.[3] ?? '';

  if (!href || !isSafeHref(href)) {
    return '';
  }

  return ` href="${escapeText(href)}"`;
};

/** Оставляет только безопасные теги: жирный, курсив, ссылки, списки, абзацы. */
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

      return `<${tag}${sanitizeAttributes(tag, attrs)}>`;
    },
  );
};

export const looksLikeHtml = (value: string): boolean => /<[a-z][\s\S]*>/i.test(value);

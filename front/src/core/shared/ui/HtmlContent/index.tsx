import clsx from 'clsx';

import { looksLikeHtml, sanitizeHtml } from '@/core/shared/lib/sanitizeHtml';

import styles from './HtmlContent.module.css';

type HtmlContentProps = {
  className?: string;
  text: string;
};

export const HtmlContent = ({ className, text }: HtmlContentProps): JSX.Element => {
  if (looksLikeHtml(text)) {
    return (
      <div
        className={clsx(styles.root, className)}
        dangerouslySetInnerHTML={{ __html: sanitizeHtml(text) }}
      />
    );
  }

  const paragraphs = text.split('\n').filter((line) => line.trim().length > 0);

  return (
    <div className={clsx(styles.root, className)}>
      {paragraphs.map((paragraph, index) => (
        <p key={`${index}-${paragraph.slice(0, 12)}`}>{paragraph}</p>
      ))}
    </div>
  );
};

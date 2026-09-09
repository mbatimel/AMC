'use client';

import { EditorContent, useEditor } from '@tiptap/react';
import clsx from 'clsx';
import { useEffect, useState } from 'react';

import { createRichTextExtensions } from '@/core/shared/lib/richTextExtensions';
import { toRichTextHtml } from '@/core/shared/lib/sanitizeHtml';

import styles from './HtmlContent.module.css';

type HtmlContentProps = {
  className?: string;
  text: string;
};

export const HtmlContent = ({ className, text }: HtmlContentProps): JSX.Element => {
  const content = toRichTextHtml(text);
  const [extensions] = useState(createRichTextExtensions);

  const editor = useEditor({
    content,
    editable: false,
    extensions,
    immediatelyRender: false,
  });

  useEffect(() => {
    if (!editor) {
      return;
    }

    editor.commands.setContent(content, { emitUpdate: false });
  }, [content, editor]);

  if (!content) {
    return <div className={clsx(styles.root, className)} />;
  }

  return <EditorContent className={clsx(styles.root, className)} editor={editor} />;
};

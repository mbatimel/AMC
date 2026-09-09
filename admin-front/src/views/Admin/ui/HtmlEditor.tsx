'use client';

import { EditorContent, useEditor, useEditorState } from '@tiptap/react';
import clsx from 'clsx';
import { useEffect, useState } from 'react';

import { isFontSizeActive, type RichTextSize } from '@/core/shared/lib/fontSize';
import { createRichTextExtensions } from '@/core/shared/lib/richTextExtensions';
import { normalizeRichTextOutput, toRichTextHtml } from '@/core/shared/lib/sanitizeHtml';

import styles from '../Admin.module.css';
import editorStyles from './HtmlEditor.module.css';

type HtmlEditorProps = {
  label: string;
  onChange: (value: string) => void;
  rows?: number;
  value: string;
};

const FONT_SIZE_PRESETS: RichTextSize[] = ['s', 'm', 'l'];

export const HtmlEditor = ({ label, onChange, value }: HtmlEditorProps): JSX.Element => {
  const [extensions] = useState(createRichTextExtensions);
  const editor = useEditor({
    content: toRichTextHtml(value),
    editorProps: {
      attributes: {
        'aria-label': label,
      },
    },
    extensions,
    immediatelyRender: false,
    onUpdate: ({ editor: current }) => {
      onChange(normalizeRichTextOutput(current.getHTML()));
    },
  });

  const toolbar = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (!current) {
        return {
          isBold: false,
          isBulletList: false,
          isItalic: false,
          isLink: false,
          isOrderedList: false,
          isSizeL: false,
          isSizeM: true,
          isSizeS: false,
        };
      }

      return {
        isBold: current.isActive('bold'),
        isBulletList: current.isActive('bulletList'),
        isItalic: current.isActive('italic'),
        isLink: current.isActive('link'),
        isOrderedList: current.isActive('orderedList'),
        isSizeL: isFontSizeActive(current, 'l'),
        isSizeM: isFontSizeActive(current, 'm'),
        isSizeS: isFontSizeActive(current, 's'),
      };
    },
  }) ?? {
    isBold: false,
    isBulletList: false,
    isItalic: false,
    isLink: false,
    isOrderedList: false,
    isSizeL: false,
    isSizeM: true,
    isSizeS: false,
  };

  useEffect(() => {
    if (!editor) {
      return;
    }

    const next = toRichTextHtml(value);
    const current = normalizeRichTextOutput(editor.getHTML());

    if (next === current || (next === '' && editor.isEmpty)) {
      return;
    }

    editor.commands.setContent(next, { emitUpdate: false });
  }, [editor, value]);

  const setLink = (): void => {
    if (!editor) {
      return;
    }

    const previous = editor.getAttributes('link').href as string | undefined;
    const href = window.prompt(
      'Ссылка (https://, mailto: или путь вида /catalog)',
      previous || 'https://',
    );

    if (href === null) {
      return;
    }

    const trimmed = href.trim();

    if (!trimmed) {
      editor.chain().focus().extendMarkRange('link').unsetLink().run();
      return;
    }

    editor.chain().focus().extendMarkRange('link').setLink({ href: trimmed }).run();
  };

  const setFontSize = (size: RichTextSize): void => {
    editor?.chain().focus().setFontSize(size).run();
  };

  return (
    <div className={clsx(styles.field)}>
      <span className={clsx(styles.label)}>{label}</span>
      <div className={clsx(editorStyles.toolbar)}>
        <button
          className={clsx(editorStyles.tool, toolbar.isBold && editorStyles.toolActive)}
          disabled={!editor}
          onClick={() => editor?.chain().focus().toggleBold().run()}
          type="button"
        >
          Жирный
        </button>
        <button
          className={clsx(editorStyles.tool, toolbar.isItalic && editorStyles.toolActive)}
          disabled={!editor}
          onClick={() => editor?.chain().focus().toggleItalic().run()}
          type="button"
        >
          Курсив
        </button>
        <button
          className={clsx(editorStyles.tool, toolbar.isLink && editorStyles.toolActive)}
          disabled={!editor}
          onClick={setLink}
          type="button"
        >
          Ссылка
        </button>
        <button
          className={clsx(editorStyles.tool, toolbar.isBulletList && editorStyles.toolActive)}
          disabled={!editor}
          onClick={() => editor?.chain().focus().toggleBulletList().run()}
          type="button"
        >
          Список
        </button>
        <button
          className={clsx(editorStyles.tool, toolbar.isOrderedList && editorStyles.toolActive)}
          disabled={!editor}
          onClick={() => editor?.chain().focus().toggleOrderedList().run()}
          type="button"
        >
          Нумерованный
        </button>
        <span className={clsx(editorStyles.toolbarDivider)} />
        {FONT_SIZE_PRESETS.map((size) => {
          const isActive =
            size === 's' ? toolbar.isSizeS : size === 'm' ? toolbar.isSizeM : toolbar.isSizeL;

          return (
            <button
              className={clsx(editorStyles.tool, isActive && editorStyles.toolActive)}
              disabled={!editor}
              key={size}
              onClick={() => setFontSize(size)}
              type="button"
            >
              {size.toUpperCase()}
            </button>
          );
        })}
      </div>
      <div className={clsx(editorStyles.shell)}>
        <EditorContent className={clsx(editorStyles.editor)} editor={editor} />
      </div>
      <p className={clsx(styles.hint)}>Enter — новый абзац, Shift+Enter — перенос строки.</p>
    </div>
  );
};

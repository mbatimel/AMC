import type { Editor } from '@tiptap/core';

import { Extension } from '@tiptap/core';

export const RICH_TEXT_SIZE_CLASS = {
  l: 'rich-text-size-l',
  m: 'rich-text-size-m',
  s: 'rich-text-size-s',
} as const;

export type RichTextSize = keyof typeof RICH_TEXT_SIZE_CLASS;

const SIZE_FROM_CLASS: Record<string, RichTextSize> = {
  [RICH_TEXT_SIZE_CLASS.l]: 'l',
  [RICH_TEXT_SIZE_CLASS.m]: 'm',
  [RICH_TEXT_SIZE_CLASS.s]: 's',
};

const isRichTextSize = (value: string): value is RichTextSize =>
  value === 's' || value === 'm' || value === 'l';

const getActiveFontSize = (editor: Editor): null | RichTextSize => {
  const value = editor.getAttributes('textStyle').fontSize as null | string | undefined;

  if (!value || !isRichTextSize(value)) {
    return null;
  }

  return value;
};

export const isFontSizeActive = (editor: Editor, size: RichTextSize): boolean => {
  const active = getActiveFontSize(editor);

  if (size === 'm') {
    return !active || active === 'm';
  }

  return active === size;
};

/** Пресеты размера шрифта S / M / L через class на textStyle. */
export const FontSize = Extension.create({
  addCommands() {
    return {
      setFontSize:
        (fontSize) =>
        ({ chain }) => {
          if (!isRichTextSize(fontSize) || fontSize === 'm') {
            return chain().setMark('textStyle', { fontSize: null }).removeEmptyTextStyle().run();
          }

          return chain().setMark('textStyle', { fontSize }).run();
        },
      unsetFontSize:
        () =>
        ({ chain }) =>
          chain().setMark('textStyle', { fontSize: null }).removeEmptyTextStyle().run(),
    };
  },

  addGlobalAttributes() {
    return [
      {
        attributes: {
          fontSize: {
            default: null,
            parseHTML: (element: HTMLElement) => {
              for (const className of Array.from(element.classList)) {
                const size = SIZE_FROM_CLASS[className];

                if (size) {
                  return size;
                }
              }

              return null;
            },
            renderHTML: (attributes: { fontSize?: null | string }) => {
              if (
                !attributes.fontSize ||
                !isRichTextSize(attributes.fontSize) ||
                attributes.fontSize === 'm'
              ) {
                return {};
              }

              return { class: RICH_TEXT_SIZE_CLASS[attributes.fontSize] };
            },
          },
        },
        types: ['textStyle'],
      },
    ];
  },

  name: 'fontSize',
});

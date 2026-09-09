import Link from '@tiptap/extension-link';
import { TextStyle } from '@tiptap/extension-text-style';
import StarterKit from '@tiptap/starter-kit';

import { FontSize } from './fontSize';

/** Фабрика расширений — новый набор на каждый редактор (нельзя шарить инстансы). */
export const createRichTextExtensions = () => [
  StarterKit.configure({
    blockquote: false,
    code: false,
    codeBlock: false,
    heading: false,
    horizontalRule: false,
    // TipTap 3 StarterKit уже тащит Link — отключаем, чтобы не дублировать.
    link: false,
    strike: false,
  }),
  TextStyle,
  FontSize,
  Link.configure({
    HTMLAttributes: {
      rel: 'noopener noreferrer',
      target: '_blank',
    },
    openOnClick: false,
    protocols: ['http', 'https', 'mailto'],
    validate: (href) => /^(https?:|mailto:|\/|#)/i.test(href.trim()),
  }),
];

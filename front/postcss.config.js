import path from 'node:path';
import { fileURLToPath } from 'node:url';
import tailwindcss from '@tailwindcss/postcss';
import postcssMixins from 'postcss-mixins';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** @type {import('postcss-load-config').Config} */
export default {
  plugins: [
    tailwindcss,
    postcssMixins({
      mixinsDir: path.join(__dirname, 'src/core/shared/styles/mixins'),
    }),
  ],
};

import grammar from './src/syntax/l.tmLanguage.json' with { type: 'json' };
import { readableCode } from './scripts/readable-code.mjs';

export default {
  themes: ['github-dark', 'github-light'],
  shiki: { langs: [grammar] },
  plugins: [readableCode],
};

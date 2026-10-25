import { defineCollection } from 'astro:content';
import { z } from 'astro/zod';
import { docsLoader, i18nLoader } from '@astrojs/starlight/loaders';
import { docsSchema, i18nSchema } from '@astrojs/starlight/schema';

export const collections = {
  i18n: defineCollection({ loader: i18nLoader(), schema: i18nSchema() }),
  docs: defineCollection({
    loader: docsLoader(),
    schema: docsSchema({ extend: z.object({
      sourcePath: z.string().optional(),
      documentRole: z.enum(['normative', 'explanatory', 'implementation', 'future']).optional(),
      generatedReference: z.boolean().optional(),
    }) }),
  }),
};

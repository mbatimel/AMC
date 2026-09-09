'use client';

import type { TermsBlock, TermsPageContent } from '@/core/shared/api/content';

import { useContent } from '@/core/entities/content';
import { InfoCard, InfoPage, InfoPageSkeleton, InfoText } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

/** Поддержка нового контракта и старого `{ title, text }` до миграции данных. */
const resolveTermsBlocks = (page: null | TermsPageContent | undefined): TermsBlock[] => {
  if (!page) {
    return [];
  }

  if (Array.isArray(page.terms) && page.terms.length > 0) {
    return page.terms;
  }

  const legacyText = (page as TermsPageContent & { text?: string }).text;

  if (typeof legacyText === 'string' && legacyText.trim().length > 0) {
    return [{ description: legacyText, title: '' }];
  }

  return [];
};

export const Terms = (): JSX.Element => {
  const { content, error, isPending } = useContent();
  const page = content?.terms;
  const blocks = resolveTermsBlocks(page);

  return (
    <Page>
      <InfoPage
        description={page?.description || undefined}
        eyebrow={page?.eyebrow || undefined}
        title={page?.title || undefined}
      >
        {isPending && !page ? (
          <InfoCard>
            <InfoPageSkeleton />
          </InfoCard>
        ) : null}

        {error && !page ? (
          <InfoCard>
            <p>{error}</p>
          </InfoCard>
        ) : null}

        {!isPending && page && blocks.length === 0 ? (
          <InfoCard>
            <p>Условия пока не опубликованы.</p>
          </InfoCard>
        ) : null}

        {blocks.map((block, index) => (
          <InfoCard
            key={`${index}-${block.title || 'block'}`}
            title={block.title.trim() ? block.title : undefined}
          >
            {block.description ? <InfoText text={block.description} /> : null}
          </InfoCard>
        ))}
      </InfoPage>
    </Page>
  );
};

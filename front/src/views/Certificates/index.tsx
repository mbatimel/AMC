'use client';

import { useContent } from '@/core/entities/content';
import {
  InfoCard,
  InfoList,
  InfoPage,
  InfoPageSkeleton,
  InfoText,
} from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

export const Certificates = (): JSX.Element => {
  const { content, error, isPending } = useContent();
  const certificates = content?.certificates;

  return (
    <Page>
      <InfoPage
        description="Документы, подтверждающие соответствие продукции ГОСТ, ТУ и системе менеджмента качества."
        eyebrow="Качество"
        title={certificates?.title ?? 'Сертификаты и лицензии'}
      >
        <InfoCard>
          {isPending && !certificates ? <InfoPageSkeleton /> : null}
          {error && !certificates ? <p>{error}</p> : null}
          {certificates ? <InfoText text={certificates.text} /> : null}
        </InfoCard>

        {certificates && certificates.items.length > 0 ? (
          <InfoCard title="Перечень документов">
            <InfoList items={certificates.items} />
          </InfoCard>
        ) : null}
      </InfoPage>
    </Page>
  );
};

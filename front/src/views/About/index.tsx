'use client';

import { useContent } from '@/core/entities/content';
import { InfoCard, InfoPage, InfoPageSkeleton, InfoText } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

export const About = (): JSX.Element => {
  const { content, error, isPending } = useContent();
  const about = content?.about;

  return (
    <Page>
      <InfoPage
        description="Производитель профессионального режущего инструмента с 1998 года."
        eyebrow="Компания"
        title={about?.title ?? 'О компании'}
      >
        <InfoCard>
          {isPending && !about ? <InfoPageSkeleton /> : null}
          {error && !about ? <p>{error}</p> : null}
          {about ? <InfoText text={about.text} /> : null}
        </InfoCard>
      </InfoPage>
    </Page>
  );
};

'use client';

import { useContent } from '@/core/entities/content';
import { InfoCard, InfoPage, InfoPageSkeleton, InfoText } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

export const Terms = (): JSX.Element => {
  const { content, error, isPending } = useContent();
  const terms = content?.terms;

  return (
    <Page>
      <InfoPage
        description="Порядок работы с оптовыми клиентами: оплата, доставка, возврат и индивидуальные условия."
        eyebrow="Сотрудничество"
        title={terms?.title ?? 'Условия работы'}
      >
        <InfoCard>
          {isPending && !terms ? <InfoPageSkeleton /> : null}
          {error && !terms ? <p>{error}</p> : null}
          {terms ? <InfoText text={terms.text} /> : null}
        </InfoCard>
      </InfoPage>
    </Page>
  );
};

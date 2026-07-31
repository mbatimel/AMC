'use client';

import { Button } from '@heroui/react';
import { useRouter } from 'next/navigation';

import { useContent } from '@/core/entities/content';
import { AppPath } from '@/core/shared/router/paths';
import {
  InfoCard,
  InfoList,
  InfoPage,
  InfoPageSkeleton,
  InfoText,
} from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

export const Promo = (): JSX.Element => {
  const router = useRouter();
  const { banners, content, error, isPending } = useContent();
  const promo = content?.promo;
  const activeBanners = (banners?.items ?? []).filter((item) => item.is_active);

  return (
    <Page>
      <InfoPage
        description="Специальные условия и сезонные предложения для оптовых клиентов портала."
        eyebrow="Выгода"
        title={promo?.title ?? 'Акции'}
      >
        <InfoCard>
          {isPending && !promo ? <InfoPageSkeleton /> : null}
          {error && !promo ? <p>{error}</p> : null}
          {promo ? <InfoText text={promo.text} /> : null}
        </InfoCard>

        {promo && promo.items.length > 0 ? (
          <InfoCard title="Действующие предложения">
            <InfoList items={promo.items} />
          </InfoCard>
        ) : null}

        {activeBanners.length > 0 ? (
          <InfoCard title="Баннеры на главной">
            <InfoList
              items={activeBanners.map((item) =>
                item.subtitle ? `${item.title} — ${item.subtitle}` : item.title,
              )}
            />
          </InfoCard>
        ) : null}

        <InfoCard>
          <Button onPress={() => router.push(AppPath.Catalog)} variant="primary">
            Перейти в каталог
          </Button>
        </InfoCard>
      </InfoPage>
    </Page>
  );
};

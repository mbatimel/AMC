import clsx from 'clsx';

import { IconSearch } from '@/core/shared/icons';

import type { UseHeaderSearchResult } from '../../model/types';

import desktopStyles from '../desktop/HeaderDesktop.module.css';
import mobileStyles from '../mobile/HeaderMobile.module.css';

type HeaderSearchFormProps = {
  search: UseHeaderSearchResult;
  variant: HeaderSearchVariant;
};

type HeaderSearchVariant = 'desktop' | 'drawer' | 'mobile';

const variantStyles: Record<HeaderSearchVariant, typeof desktopStyles> = {
  desktop: desktopStyles,
  drawer: mobileStyles,
  mobile: mobileStyles,
};

export const HeaderSearchForm = ({ search, variant }: HeaderSearchFormProps): JSX.Element => {
  const styles = variantStyles[variant];
  const placeholder = variant === 'desktop' ? search.desktopPlaceholder : search.mobilePlaceholder;
  const isIconOnlySubmit = variant === 'mobile' || variant === 'drawer';
  const formClassName = clsx(
    styles.search,
    variant === 'mobile' && mobileStyles.searchMobile,
    variant === 'drawer' && mobileStyles.searchDrawer,
  );

  return (
    <form className={formClassName} onSubmit={search.onSearchSubmit} role="search">
      {variant === 'drawer' && <IconSearch className={clsx(mobileStyles.searchDrawerIcon)} height={16} width={16} />}
      <input
        className={clsx(styles.searchInput)}
        name="query"
        onChange={(event) => search.onQueryChange(event.target.value)}
        placeholder={placeholder}
        type="search"
        value={search.query}
      />
      <button
        aria-label={isIconOnlySubmit ? 'Найти' : undefined}
        className={clsx(styles.searchButton, isIconOnlySubmit && mobileStyles.searchButtonIcon)}
        type="submit"
      >
        <IconSearch height={16} width={16} />
        {!isIconOnlySubmit && <span>Найти</span>}
      </button>
    </form>
  );
};

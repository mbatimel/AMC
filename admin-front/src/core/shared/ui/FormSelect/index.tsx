'use client';

import type { Key } from '@heroui/react';

import {
  Autocomplete,
  EmptyState,
  Header,
  Label,
  ListBox,
  SearchField,
  Tag,
  TagGroup,
  useFilter,
} from '@heroui/react';
import clsx from 'clsx';
import { useMemo, useState } from 'react';

import { IconChevronDown } from '@/core/shared/icons';
import { useDebouncedValue } from '@/core/shared/lib/useDebouncedValue';

import styles from './FormSelect.module.css';

const SEARCH_DEBOUNCE_MS = 250;

export type FormSelectMultipleProps = FormSelectBaseProps & {
  onChange: (value: string[]) => void;
  selectionMode: 'multiple';
  value: string[];
};

export type FormSelectOption = {
  label: string;
  value: string;
};

export type FormSelectOptionGroup = {
  children?: FormSelectOptionGroup[];
  id: string;
  label: string;
  options?: FormSelectOption[];
};

export type FormSelectProps = FormSelectMultipleProps | FormSelectSingleProps;

export type FormSelectSingleProps = FormSelectBaseProps & {
  onChange: (value: string) => void;
  selectionMode?: 'single';
  value: string;
};

type FormSelectBaseProps = {
  ariaLabel: string;
  className?: string;
  error?: string;
  groups?: FormSelectOptionGroup[];
  isDisabled?: boolean;
  label?: string;
  options?: FormSelectOption[];
  placeholder?: string;
};

const toStringKeys = (keys: Iterable<Key>): string[] =>
  Array.from(keys).flatMap((key) => (typeof key === 'string' ? [key] : []));

export const flattenFormSelectGroups = (groups: FormSelectOptionGroup[]): FormSelectOption[] =>
  groups.flatMap((group) => [
    ...(group.options ?? []),
    ...flattenFormSelectGroups(group.children ?? []),
  ]);

const resolveOptions = (
  options: FormSelectOption[] | undefined,
  groups: FormSelectOptionGroup[] | undefined,
): FormSelectOption[] => options ?? flattenFormSelectGroups(groups ?? []);

const groupHasMatch = (
  group: FormSelectOptionGroup,
  filterText: string,
  contains: (text: string, filter: string) => boolean,
): boolean => {
  if ((group.options ?? []).some((option) => contains(option.label, filterText))) {
    return true;
  }

  return (group.children ?? []).some((child) => groupHasMatch(child, filterText, contains));
};

/** Секции и items — прямые потомки ListBox; свёртка только через CSS, иначе RAC collection пустеет. */
const renderGroupSections = ({
  contains,
  depth,
  expandedIds,
  filterText,
  group,
  onToggle,
  parentExpanded,
}: {
  contains: (text: string, filter: string) => boolean;
  depth: number;
  expandedIds: ReadonlySet<string>;
  filterText: string;
  group: FormSelectOptionGroup;
  onToggle: (id: string) => void;
  parentExpanded: boolean;
}): JSX.Element[] => {
  const isSearching = filterText.trim().length > 0;
  const allOptions = group.options ?? [];
  const matchedOptions = isSearching
    ? allOptions.filter((option) => contains(option.label, filterText))
    : allOptions;
  const matchedChildren = (group.children ?? []).filter((child) =>
    isSearching ? groupHasMatch(child, filterText, contains) : true,
  );

  if (matchedOptions.length === 0 && matchedChildren.length === 0) {
    return [];
  }

  const isExpanded = isSearching || expandedIds.has(group.id);
  const showHeader =
    parentExpanded && (!isSearching || matchedOptions.length > 0 || matchedChildren.length > 0);
  const showItems = parentExpanded && isExpanded;
  const totalCount = allOptions.length + flattenFormSelectGroups(group.children ?? []).length;

  return [
    <ListBox.Section id={group.id} key={group.id}>
      <Header className={clsx(styles.groupHeader, !showHeader && styles.itemCollapsed)}>
        <button
          aria-expanded={isExpanded}
          aria-label={`${isExpanded ? 'Свернуть' : 'Развернуть'} категорию ${group.label}`}
          className={clsx(styles.groupToggle)}
          onClick={(event) => {
            event.preventDefault();
            event.stopPropagation();
            if (!isSearching) {
              onToggle(group.id);
            }
          }}
          style={{ paddingInlineStart: `${8 + depth * 12}px` }}
          type="button"
        >
          <IconChevronDown
            className={clsx(styles.groupChevron, isExpanded && styles.groupChevronOpen)}
            height={12}
            width={12}
          />
          <span className={clsx(styles.groupLabel)}>{group.label}</span>
          <span className={clsx(styles.groupCount)}>{totalCount}</span>
        </button>
      </Header>
      {matchedOptions.map((option) => (
        <ListBox.Item
          className={clsx(styles.optionItem, !showItems && styles.itemCollapsed)}
          id={option.value}
          key={option.value}
          textValue={option.label}
        >
          {option.label}
          <ListBox.ItemIndicator />
        </ListBox.Item>
      ))}
    </ListBox.Section>,
    ...matchedChildren.flatMap((child) =>
      renderGroupSections({
        contains,
        depth: depth + 1,
        expandedIds,
        filterText,
        group: child,
        onToggle,
        parentExpanded: showItems,
      }),
    ),
  ];
};

const OptionsList = ({
  ariaLabel,
  contains,
  filterText,
  groups,
  options,
}: {
  ariaLabel: string;
  contains: (text: string, filter: string) => boolean;
  filterText: string;
  groups?: FormSelectOptionGroup[];
  options: FormSelectOption[];
}): JSX.Element => {
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => new Set());

  const onToggle = (id: string): void => {
    setExpandedIds((previous) => {
      const next = new Set(previous);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  if (!groups || groups.length === 0) {
    return (
      <ListBox
        aria-label={ariaLabel}
        className={clsx(styles.list)}
        renderEmptyState={() => <EmptyState>Ничего не найдено</EmptyState>}
      >
        {options.map((option) => (
          <ListBox.Item
            className={clsx(styles.optionItem)}
            id={option.value}
            key={option.value}
            textValue={option.label}
          >
            {option.label}
            <ListBox.ItemIndicator />
          </ListBox.Item>
        ))}
      </ListBox>
    );
  }

  return (
    <ListBox
      aria-label={ariaLabel}
      className={clsx(styles.list)}
      renderEmptyState={() => <EmptyState>Ничего не найдено</EmptyState>}
    >
      {groups.flatMap((group) =>
        renderGroupSections({
          contains,
          depth: 0,
          expandedIds,
          filterText,
          group,
          onToggle,
          parentExpanded: true,
        }),
      )}
    </ListBox>
  );
};

const FormSelectShell = ({
  ariaLabel,
  className,
  error,
  groups,
  isDisabled,
  label,
  onChange,
  onClear,
  options,
  placeholder,
  selectionMode,
  trigger,
  value,
}: {
  ariaLabel: string;
  className?: string;
  error?: string;
  groups?: FormSelectOptionGroup[];
  isDisabled?: boolean;
  label?: string;
  onChange: (keys: Key | Key[] | null) => void;
  onClear: () => void;
  options: FormSelectOption[];
  placeholder: string;
  selectionMode: 'multiple' | 'single';
  trigger: JSX.Element;
  value: Key | Key[] | null;
}): JSX.Element => {
  const { contains } = useFilter({ sensitivity: 'base' });
  const [filterText, setFilterText] = useState('');
  const debouncedFilterText = useDebouncedValue(
    filterText,
    filterText === '' ? 0 : SEARCH_DEBOUNCE_MS,
  );

  return (
    <div className={clsx(styles.root, className)}>
      <Autocomplete
        aria-label={ariaLabel}
        className={clsx(styles.control, Boolean(error) && styles.controlInvalid)}
        fullWidth
        isDisabled={isDisabled}
        isInvalid={Boolean(error)}
        onChange={onChange}
        onClear={onClear}
        onOpenChange={(isOpen) => {
          if (!isOpen) {
            setFilterText('');
          }
        }}
        placeholder={placeholder}
        selectionMode={selectionMode}
        value={value}
      >
        {label ? <Label className={clsx(styles.label)}>{label}</Label> : null}
        {trigger}
        <Autocomplete.Popover className={clsx(styles.popover)}>
          <Autocomplete.Filter filter={(textValue) => contains(textValue, debouncedFilterText)}>
            <SearchField
              aria-label={`Поиск: ${ariaLabel}`}
              autoFocus
              className={clsx(styles.search)}
              name="search"
              onChange={setFilterText}
              value={filterText}
              variant="secondary"
            >
              <SearchField.Group>
                <SearchField.SearchIcon />
                <SearchField.Input placeholder="Поиск…" />
                <SearchField.ClearButton aria-label="Очистить поиск" />
              </SearchField.Group>
            </SearchField>
            <OptionsList
              ariaLabel={ariaLabel}
              contains={contains}
              filterText={debouncedFilterText}
              groups={groups}
              key={groups?.map((group) => group.id).join(',') || 'flat'}
              options={options}
            />
          </Autocomplete.Filter>
        </Autocomplete.Popover>
      </Autocomplete>
      {error ? <p className={clsx(styles.error)}>{error}</p> : null}
    </div>
  );
};

const FormSelectSingle = ({
  ariaLabel,
  className,
  error,
  groups,
  isDisabled,
  label,
  onChange,
  options,
  placeholder = 'Не выбрано',
  value,
}: FormSelectSingleProps): JSX.Element => {
  const resolvedOptions = useMemo(() => resolveOptions(options, groups), [groups, options]);

  return (
    <FormSelectShell
      ariaLabel={ariaLabel}
      className={className}
      error={error}
      groups={groups}
      isDisabled={isDisabled || (resolvedOptions.length === 0 && !value)}
      label={label}
      onChange={(keys) => onChange(typeof keys === 'string' ? keys : '')}
      onClear={() => onChange('')}
      options={resolvedOptions}
      placeholder={placeholder}
      selectionMode="single"
      trigger={
        <Autocomplete.Trigger className={clsx(styles.trigger)}>
          <Autocomplete.Value />
          <Autocomplete.ClearButton aria-label="Очистить выбор" />
          <Autocomplete.Indicator />
        </Autocomplete.Trigger>
      }
      value={value || null}
    />
  );
};

const FormSelectMultiple = ({
  ariaLabel,
  className,
  error,
  groups,
  isDisabled,
  label,
  onChange,
  options,
  placeholder = 'Не выбрано',
  value,
}: FormSelectMultipleProps): JSX.Element => {
  const resolvedOptions = useMemo(() => resolveOptions(options, groups), [groups, options]);
  const optionByValue = useMemo(
    () => new Map(resolvedOptions.map((option) => [option.value, option])),
    [resolvedOptions],
  );

  return (
    <FormSelectShell
      ariaLabel={ariaLabel}
      className={className}
      error={error}
      groups={groups}
      isDisabled={isDisabled || (resolvedOptions.length === 0 && value.length === 0)}
      label={label}
      onChange={(keys) => onChange(Array.isArray(keys) ? toStringKeys(keys) : [])}
      onClear={() => onChange([])}
      options={resolvedOptions}
      placeholder={placeholder}
      selectionMode="multiple"
      trigger={
        <Autocomplete.Trigger className={clsx(styles.trigger, styles.triggerMultiple)}>
          <Autocomplete.Value>
            {({ defaultChildren, isPlaceholder, state }) => {
              if (isPlaceholder || state.selectedItems.length === 0) {
                return defaultChildren;
              }

              return (
                <TagGroup
                  aria-label={`Выбрано: ${ariaLabel}`}
                  onRemove={(keys) => {
                    const removed = new Set(toStringKeys(keys));
                    onChange(value.filter((id) => !removed.has(id)));
                  }}
                  size="sm"
                >
                  <TagGroup.List>
                    {state.selectedItems.map((item) => {
                      const id = String(item.key);
                      const option = optionByValue.get(id);

                      return (
                        <Tag id={id} key={id}>
                          {option?.label ?? id}
                        </Tag>
                      );
                    })}
                  </TagGroup.List>
                </TagGroup>
              );
            }}
          </Autocomplete.Value>
          <Autocomplete.ClearButton aria-label="Очистить выбор" />
          <Autocomplete.Indicator />
        </Autocomplete.Trigger>
      }
      value={value}
    />
  );
};

/**
 * Поиск и выбор из списка на базе HeroUI `Autocomplete`.
 * Опционально — дерево групп (`groups`) с раскрывашками.
 */
export const FormSelect = (props: FormSelectProps): JSX.Element =>
  props.selectionMode === 'multiple' ? (
    <FormSelectMultiple {...props} />
  ) : (
    <FormSelectSingle {...props} />
  );

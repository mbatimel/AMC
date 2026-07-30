'use client';

import {
  Dropdown,
  DropdownItem,
  DropdownMenu,
  DropdownPopover,
  DropdownTrigger,
} from '@heroui/react';
import clsx from 'clsx';

import type { CityItem } from '@/core/shared/api/orders';

import { IconChevronDown, IconLocation } from '@/core/shared/icons';

import styles from './HeaderCitySelect.module.css';

type HeaderCitySelectProps = {
  cities: CityItem[];
  isPending: boolean;
  onCitySelect: (cityId: string) => void;
  selectedCityId: null | string;
  selectedCityName: string;
};

export const HeaderCitySelect = ({
  cities,
  isPending,
  onCitySelect,
  selectedCityId,
  selectedCityName,
}: HeaderCitySelectProps): JSX.Element => {
  return (
    <Dropdown>
      <DropdownTrigger
        aria-label="Выбрать город"
        className={clsx(styles.trigger)}
        isDisabled={isPending || cities.length === 0}
      >
        <IconLocation className={clsx(styles.icon)} height={14} width={14} />
        <span>{selectedCityName}</span>
        <IconChevronDown className={clsx(styles.icon)} height={12} width={12} />
      </DropdownTrigger>

      <DropdownPopover className={clsx(styles.popover)}>
        <DropdownMenu
          aria-label="Города"
          onSelectionChange={(keys) => {
            const cityId = Array.from(keys)[0];

            if (typeof cityId === 'string' && cityId.length > 0) {
              onCitySelect(cityId);
            }
          }}
          selectedKeys={selectedCityId ? [selectedCityId] : []}
          selectionMode="single"
        >
          {cities.map((city) => (
            <DropdownItem id={city.id} key={city.id} textValue={city.city}>
              {city.city}
            </DropdownItem>
          ))}
        </DropdownMenu>
      </DropdownPopover>
    </Dropdown>
  );
};

import type { JSX } from 'react';

export type Icon = (props: IconProps) => JSX.Element;

export type IconProps = {
  className?: string;
  currentColor?: string;
  height?: number;
  width?: number;
};

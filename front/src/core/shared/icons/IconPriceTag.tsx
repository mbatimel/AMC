import type { IconProps } from './types';

export const IconPriceTag = ({
  className,
  currentColor = 'currentColor',
  height = 12,
  width = 12,
}: IconProps): JSX.Element => {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      height={height}
      viewBox="0 0 12 12"
      width={width}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M6.293 1.293C6.10551 1.10545 5.85119 1.00006 5.586 1H2C1.73478 1 1.48043 1.10536 1.29289 1.29289C1.10536 1.48043 1 1.73478 1 2V5.586C1.00006 5.85119 1.10545 6.10551 1.293 6.293L5.645 10.645C5.87226 10.8708 6.17962 10.9976 6.5 10.9976C6.82038 10.9976 7.12774 10.8708 7.355 10.645L10.645 7.355C10.8708 7.12774 10.9976 6.82038 10.9976 6.5C10.9976 6.17962 10.8708 5.87226 10.645 5.645L6.293 1.293Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M3.75 4C3.88807 4 4 3.88807 4 3.75C4 3.61193 3.88807 3.5 3.75 3.5C3.61193 3.5 3.5 3.61193 3.5 3.75C3.5 3.88807 3.61193 4 3.75 4Z"
        fill={currentColor}
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};

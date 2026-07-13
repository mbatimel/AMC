import type { IconProps } from './types';

export const IconChevronDown = ({
  className,
  currentColor = 'currentColor',
  height = 16,
  width = 16,
}: IconProps): JSX.Element => {
  return (
    <svg
      className={className}
      fill="none"
      height={height}
      viewBox="0 0 12 12"
      width={width}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M3 4.5L6 7.5L9 4.5"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeOpacity="0.92"
      />
    </svg>
  );
};

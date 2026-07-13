import type { IconProps } from './types';

export const IconMenu = ({
  className,
  currentColor = 'currentColor',
  height = 16,
  width = 16,
}: IconProps): JSX.Element => {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      height={height}
      viewBox="0 0 16 16"
      width={width}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M2.667 4h10.666M2.667 8h10.666M2.667 12h10.666" stroke={currentColor} strokeLinecap="round" strokeWidth="1.5" />
    </svg>
  );
};

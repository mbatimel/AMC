import type { IconProps } from './types';

export const IconClose = ({
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
      <path d="m4 4 8 8M12 4 4 12" stroke={currentColor} strokeLinecap="round" strokeWidth="1.5" />
    </svg>
  );
};

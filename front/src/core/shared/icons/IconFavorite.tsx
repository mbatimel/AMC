import type { IconProps } from './types';

export const IconFavorite = ({
  className,
  currentColor = 'currentColor',
  height = 18,
  width = 18,
}: IconProps): JSX.Element => {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      height={height}
      viewBox="0 0 18 18"
      width={width}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M9 2.625L10.854 6.804L15.375 7.317L12.018 10.386L12.945 14.875L9 12.623L5.055 14.875L5.982 10.386L2.625 7.317L7.146 6.804L9 2.625Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
    </svg>
  );
};

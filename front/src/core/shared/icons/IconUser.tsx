import type { IconProps } from './types';

export const IconUser = ({
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
      viewBox="0 0 16 16"
      width={width}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M8.00002 8.66667C9.84097 8.66667 11.3334 7.17428 11.3334 5.33333C11.3334 3.49238 9.84097 2 8.00002 2C6.15907 2 4.66669 3.49238 4.66669 5.33333C4.66669 7.17428 6.15907 8.66667 8.00002 8.66667Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M13.3334 14C13.3334 12.5855 12.7715 11.2289 11.7713 10.2287C10.7711 9.22853 9.41451 8.66663 8.00002 8.66663C6.58553 8.66663 5.22898 9.22853 4.22878 10.2287C3.22859 11.2289 2.66669 12.5855 2.66669 14"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
    </svg>
  );
};

import type { IconProps } from './types';

export const IconSearch = ({
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
      viewBox="0 0 14 14"
      width={width}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M12.25 12.2501L9.71832 9.71838"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.16667"
      />
      <path
        d="M6.41667 11.0833C8.994 11.0833 11.0833 8.994 11.0833 6.41667C11.0833 3.83934 8.994 1.75 6.41667 1.75C3.83934 1.75 1.75 3.83934 1.75 6.41667C1.75 8.994 3.83934 11.0833 6.41667 11.0833Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.16667"
      />
    </svg>
  );
};

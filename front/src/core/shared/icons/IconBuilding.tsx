import type { IconProps } from './types';

export const IconBuilding = ({
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
        d="M2 14H14"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M3.33333 14V4.66667L8 2L12.6667 4.66667V14"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M6 6.66663H6.66667"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M9.33333 6.66663H10"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M6 9.33337H6.66667"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M9.33333 9.33337H10"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M6.66667 14V11.3334H9.33333V14"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
    </svg>
  );
};

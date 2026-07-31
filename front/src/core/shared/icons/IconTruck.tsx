import type { IconProps } from './types';

export const IconTruck = ({
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
        d="M11.25 3H2.25V11.25H11.25V3Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M11.25 6H13.5L15.75 8.25V11.25H11.25V6Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M4.125 14.25C4.74632 14.25 5.25 13.7463 5.25 13.125C5.25 12.5037 4.74632 12 4.125 12C3.50368 12 3 12.5037 3 13.125C3 13.7463 3.50368 14.25 4.125 14.25Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M13.875 14.25C14.4963 14.25 15 13.7463 15 13.125C15 12.5037 14.4963 12 13.875 12C13.2537 12 12.75 12.5037 12.75 13.125C12.75 13.7463 13.2537 14.25 13.875 14.25Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
    </svg>
  );
};

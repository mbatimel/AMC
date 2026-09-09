import type { IconProps } from './types';

/** Замок — восстановление пароля. */
export const IconKey = ({
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
      <rect
        height="7.333"
        rx="1.333"
        stroke={currentColor}
        strokeWidth="1.33333"
        width="10"
        x="3"
        y="7.333"
      />
      <path
        d="M5.333 7.333V5.333a2.667 2.667 0 0 1 5.334 0v2"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
    </svg>
  );
};

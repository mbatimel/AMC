import type { IconProps } from './types';

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
      <path
        d="M10.6667 5.33337C11.7712 5.33337 12.6667 6.2288 12.6667 7.33337C12.6667 8.43794 11.7712 9.33337 10.6667 9.33337C9.56209 9.33337 8.66666 8.43794 8.66666 7.33337C8.66666 6.2288 9.56209 5.33337 10.6667 5.33337Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M10.6667 9.33337V14"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M8.66666 12H12.6667"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M2 7.33337C2 4.38785 4.38781 2 7.33333 2C9.02266 2 10.5307 2.78385 11.5 4"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
    </svg>
  );
};

import type { IconProps } from './types';

export const IconBrands = ({
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
        d="M11.6076 9.66797L12.7438 16.0625C12.7565 16.1378 12.746 16.2152 12.7135 16.2843C12.6811 16.3534 12.6283 16.411 12.5623 16.4493C12.4962 16.4876 12.42 16.5049 12.3439 16.4988C12.2678 16.4926 12.1954 16.4634 12.1363 16.415L9.45131 14.3997C9.32169 14.3029 9.16424 14.2506 9.00244 14.2506C8.84064 14.2506 8.68318 14.3029 8.55356 14.3997L5.86406 16.4142C5.80505 16.4625 5.73271 16.4917 5.65669 16.4979C5.58066 16.504 5.50457 16.4868 5.43856 16.4486C5.37255 16.4104 5.31977 16.353 5.28725 16.284C5.25473 16.215 5.24403 16.1377 5.25656 16.0625L6.39206 9.66797"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M9 10.5C11.4853 10.5 13.5 8.48528 13.5 6C13.5 3.51472 11.4853 1.5 9 1.5C6.51472 1.5 4.5 3.51472 4.5 6C4.5 8.48528 6.51472 10.5 9 10.5Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
    </svg>
  );
};

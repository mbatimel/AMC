import type { IconProps } from './types';

export const IconCart = ({
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
      <g clipPath="url(#clip0_129_11559)">
        <path
          d="M6 16.5C6.41421 16.5 6.75 16.1642 6.75 15.75C6.75 15.3358 6.41421 15 6 15C5.58579 15 5.25 15.3358 5.25 15.75C5.25 16.1642 5.58579 16.5 6 16.5Z"
          stroke={currentColor}
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="1.5"
        />
        <path
          d="M14.25 16.5C14.6642 16.5 15 16.1642 15 15.75C15 15.3358 14.6642 15 14.25 15C13.8358 15 13.5 15.3358 13.5 15.75C13.5 16.1642 13.8358 16.5 14.25 16.5Z"
          stroke={currentColor}
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="1.5"
        />
        <path
          d="M1.5376 1.53711H3.0376L5.0326 10.8521C5.10578 11.1933 5.2956 11.4982 5.56938 11.7145C5.84316 11.9308 6.18378 12.0449 6.5326 12.0371H13.8676C14.209 12.0366 14.54 11.9196 14.8059 11.7055C15.0718 11.4914 15.2567 11.193 15.3301 10.8596L16.5676 5.28711H3.8401"
          stroke={currentColor}
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="1.5"
        />
      </g>
      <defs>
        <clipPath id="clip0_129_11559">
          <rect fill="white" height="18" width="18" />
        </clipPath>
      </defs>
    </svg>
  );
};

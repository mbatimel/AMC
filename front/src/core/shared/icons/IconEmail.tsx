import type { IconProps } from './types';

export const IconEmail = ({
  className,
  currentColor = 'currentColor',
  height = 13,
  width = 13,
}: IconProps): JSX.Element => {
  return (
    <svg
      className={className}
      fill="none"
      height={height}
      viewBox="0 0 13 13"
      width={width}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M11.9167 3.79163L7.04654 6.89375C6.88127 6.98974 6.69355 7.0403 6.50243 7.0403C6.31131 7.0403 6.12359 6.98974 5.95833 6.89375L1.08333 3.79163"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeOpacity="0.72"
        strokeWidth="1.08333"
      />
      <path
        d="M10.8333 2.16663H2.16666C1.56835 2.16663 1.08333 2.65165 1.08333 3.24996V9.74996C1.08333 10.3483 1.56835 10.8333 2.16666 10.8333H10.8333C11.4316 10.8333 11.9167 10.3483 11.9167 9.74996V3.24996C11.9167 2.65165 11.4316 2.16663 10.8333 2.16663Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeOpacity="0.72"
        strokeWidth="1.08333"
      />
    </svg>
  );
};

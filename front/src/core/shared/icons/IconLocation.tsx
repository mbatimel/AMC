import type { IconProps } from './types';

export const IconLocation = ({
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
        d="M6.82553 11.8078C7.83303 10.9379 10.8333 8.12125 10.8333 5.41671C10.8333 4.26744 10.3768 3.16524 9.56412 2.35258C8.75146 1.53992 7.64926 1.08337 6.49999 1.08337C5.35072 1.08337 4.24852 1.53992 3.43586 2.35258C2.6232 3.16524 2.16666 4.26744 2.16666 5.41671C2.16666 8.12125 5.16695 10.9379 6.17445 11.8078C6.26831 11.8784 6.38256 11.9166 6.49999 11.9166C6.61742 11.9166 6.73167 11.8784 6.82553 11.8078Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeOpacity="0.92"
        strokeWidth="1.08333"
      />
      <path
        d="M6.5 7.04163C7.39746 7.04163 8.125 6.31409 8.125 5.41663C8.125 4.51916 7.39746 3.79163 6.5 3.79163C5.60254 3.79163 4.875 4.51916 4.875 5.41663C4.875 6.31409 5.60254 7.04163 6.5 7.04163Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeOpacity="0.92"
        strokeWidth="1.08333"
      />
    </svg>
  );
};

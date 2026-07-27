import type { IconProps } from './types';

export const IconMail = ({
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
        d="M2.66667 3.33337H13.3333C14.0667 3.33337 14.6667 3.93337 14.6667 4.66671V11.3334C14.6667 12.0667 14.0667 12.6667 13.3333 12.6667H2.66667C1.93333 12.6667 1.33333 12.0667 1.33333 11.3334V4.66671C1.33333 3.93337 1.93333 3.33337 2.66667 3.33337Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
      <path
        d="M14.6667 4.66663L8 8.66663L1.33333 4.66663"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.33333"
      />
    </svg>
  );
};

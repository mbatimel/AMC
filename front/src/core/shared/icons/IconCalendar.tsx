import type { IconProps } from './types';

/** Иконка календаря для срока акции. */
export const IconCalendar = ({
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
        d="M5.25 2.25V4.5M12.75 2.25V4.5M2.625 6.75H15.375M3.75 3.375H14.25C14.8713 3.375 15.375 3.87868 15.375 4.5V14.25C15.375 14.8713 14.8713 15.375 14.25 15.375H3.75C3.12868 15.375 2.625 14.8713 2.625 14.25V4.5C2.625 3.87868 3.12868 3.375 3.75 3.375Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
    </svg>
  );
};

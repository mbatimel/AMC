import type { IconProps } from './types';

export const IconTerms = ({
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
        d="M11.25 9H7.5"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M11.25 6H7.5"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M14.25 12.75V3.75C14.25 3.35218 14.092 2.97064 13.8107 2.68934C13.5294 2.40804 13.1478 2.25 12.75 2.25H3"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M6 15.75H15C15.3978 15.75 15.7794 15.592 16.0607 15.3107C16.342 15.0294 16.5 14.6478 16.5 14.25V13.5C16.5 13.3011 16.421 13.1103 16.2803 12.9697C16.1397 12.829 15.9489 12.75 15.75 12.75H8.25C8.05109 12.75 7.86032 12.829 7.71967 12.9697C7.57902 13.1103 7.5 13.3011 7.5 13.5V14.25C7.5 14.6478 7.34196 15.0294 7.06066 15.3107C6.77936 15.592 6.39782 15.75 6 15.75ZM6 15.75C5.60218 15.75 5.22064 15.592 4.93934 15.3107C4.65804 15.0294 4.5 14.6478 4.5 14.25V3.75C4.5 3.35218 4.34196 2.97064 4.06066 2.68934C3.77936 2.40804 3.39782 2.25 3 2.25C2.60218 2.25 2.22064 2.40804 1.93934 2.68934C1.65804 2.97064 1.5 3.35218 1.5 3.75V5.25C1.5 5.44891 1.57902 5.63968 1.71967 5.78033C1.86032 5.92098 2.05109 6 2.25 6H4.5"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
    </svg>
  );
};

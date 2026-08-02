import type { IconProps } from './types';

/** Иконка упаковки для количества товаров. */
export const IconPackage = ({
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
        d="M12.1875 6.9375L5.8125 3.2625M15.375 12V6C15.3747 5.80289 15.3228 5.60955 15.2244 5.43915C15.1261 5.26876 14.9847 5.12713 14.8144 5.02837L9.56437 2.02837C9.3938 1.92945 9.2001 1.87732 9.00281 1.87732C8.80553 1.87732 8.61182 1.92945 8.44125 2.02837L3.19125 5.02837C3.02095 5.12713 2.87955 5.26876 2.7812 5.43915C2.68285 5.60955 2.63092 5.80289 2.63062 6V12C2.63092 12.1971 2.68285 12.3905 2.7812 12.5608C2.87955 12.7312 3.02095 12.8729 3.19125 12.9716L8.44125 15.9716C8.61182 16.0705 8.80553 16.1227 9.00281 16.1227C9.2001 16.1227 9.3938 16.0705 9.56437 15.9716L14.8144 12.9716C14.9847 12.8729 15.1261 12.7312 15.2244 12.5608C15.3228 12.3905 15.3747 12.1971 15.375 12Z"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
      <path
        d="M2.7825 5.415L9 9.0075L15.2175 5.415M9 16.2V9"
        stroke={currentColor}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      />
    </svg>
  );
};

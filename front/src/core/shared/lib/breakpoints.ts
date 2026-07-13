export const BREAKPOINT_MOBILE_MAX = 680;

export const BREAKPOINT_TABLET_MIN = 681;

export const BREAKPOINT_TABLET_MAX = 1074;

export const BREAKPOINT_DESKTOP_MIN = 1075;

/** Mobile + tablet layout (compact header/footer) */
export const getMobileLayoutMediaQuery = (): string => `(width <= ${BREAKPOINT_TABLET_MAX}px)`;

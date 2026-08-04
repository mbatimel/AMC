export const toTelHref = (phone: string): string => `tel:${phone.replace(/[^\d+]/g, '')}`;

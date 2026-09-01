import type { RegisterIpPayload } from '@/core/shared/api/auth';

import { readFormString } from '@/core/shared/lib/readFormString';

const readOptional = (formData: FormData, key: string): string | undefined => {
  const value = readFormString(formData, key).trim();

  return value.length > 0 ? value : undefined;
};

export const buildRegisterPayload = (formData: FormData): RegisterIpPayload => {
  return {
    actualAddress: readOptional(formData, 'actualAddress'),
    additionalPhone: readOptional(formData, 'phoneAdditional'),
    bankAccount: readOptional(formData, 'bankAccount'),
    bankBik: readOptional(formData, 'bik'),
    bankName: readOptional(formData, 'bankName'),
    correspondentAccount: readOptional(formData, 'corrAccount'),
    directorFullName: readOptional(formData, 'directorFullName'),
    directorPosition: readOptional(formData, 'directorPosition'),
    email: readFormString(formData, 'email').trim(),
    fullName: readOptional(formData, 'fullCompanyName'),
    inn: readOptional(formData, 'inn'),
    kpp: readOptional(formData, 'kpp'),
    legalAddress: readOptional(formData, 'legalAddress'),
    ogrn: readOptional(formData, 'ogrn'),
    okved: readOptional(formData, 'okved'),
    password: readFormString(formData, 'password'),
    phone: readOptional(formData, 'phone'),
    shortName: readOptional(formData, 'shortCompanyName'),
    taxSystem: readOptional(formData, 'taxSystem'),
    website: readOptional(formData, 'website'),
  };
};

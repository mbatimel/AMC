import type { RegisterIndividualPayload, RegisterIpPayload } from '@/core/shared/api/auth';

import { readFormString } from '@/core/shared/lib/readFormString';

export enum RegisterType {
  Individual = 'individual',
  Organization = 'organization',
}

const readOptional = (formData: FormData, key: string): string | undefined => {
  const value = readFormString(formData, key).trim();

  return value.length > 0 ? value : undefined;
};

export const buildRegisterIpPayload = (formData: FormData): RegisterIpPayload => {
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

export const buildRegisterIndividualPayload = (formData: FormData): RegisterIndividualPayload => {
  return {
    city: readFormString(formData, 'city').trim(),
    deliveryAddress: readFormString(formData, 'deliveryAddress').trim(),
    email: readFormString(formData, 'email').trim(),
    fio: readFormString(formData, 'fullName').trim(),
    inn: readOptional(formData, 'inn'),
    password: readFormString(formData, 'password'),
    phone: readFormString(formData, 'phone').trim(),
  };
};

export type RegisterPayload =
  | { data: RegisterIndividualPayload; type: RegisterType.Individual }
  | { data: RegisterIpPayload; type: RegisterType.Organization };

export const buildRegisterPayload = (
  formData: FormData,
  registerType: RegisterType,
): RegisterPayload => {
  if (registerType === RegisterType.Organization) {
    return {
      data: buildRegisterIpPayload(formData),
      type: RegisterType.Organization,
    };
  }

  return {
    data: buildRegisterIndividualPayload(formData),
    type: RegisterType.Individual,
  };
};

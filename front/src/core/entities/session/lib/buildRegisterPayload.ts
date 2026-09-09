import type { RegisterIpPayload } from '@/core/shared/api/auth';

import { readFormString } from '@/core/shared/lib/readFormString';

const readRequiredFile = (formData: FormData, key: string): File => {
  const value = formData.get(key);

  if (!(value instanceof File) || value.size === 0) {
    throw new Error('Прикрепите файл с реквизитами');
  }

  return value;
};

export const buildRegisterPayload = (formData: FormData): RegisterIpPayload => {
  return {
    directorFullName: readFormString(formData, 'directorFullName').trim(),
    email: readFormString(formData, 'email').trim(),
    inn: readFormString(formData, 'inn').trim(),
    password: readFormString(formData, 'password'),
    phone: readFormString(formData, 'phone').trim(),
    requisitesFile: readRequiredFile(formData, 'requisitesFile'),
    shortName: readFormString(formData, 'shortCompanyName').trim(),
  };
};

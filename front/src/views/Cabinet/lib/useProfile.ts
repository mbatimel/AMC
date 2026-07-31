'use client';

import { type FormEvent, useState } from 'react';

import type { Profile } from '@/core/shared/api/profile';

import {
  formatPhoneDisplay,
  formatPhoneInput,
  normalizePhone,
  validateEmail,
  validatePhone,
  validateRequired,
} from '@/core/shared/lib/validateContact';

export type ProfileFormSavePayload = {
  email: string;
  firstName: string;
  lastName: string;
  middleName: string;
  phone: string;
};

type UseProfileFormParams = {
  onSave: (payload: ProfileFormSavePayload) => void;
  profile: Profile;
};

export const useProfileForm = ({
  onSave,
  profile,
}: UseProfileFormParams): {
  email: string;
  firstName: string;
  handleSubmit: (event: FormEvent<HTMLFormElement>) => void;
  lastName: string;
  middleName: string;
  phone: string;
  setEmail: (value: string) => void;
  setFirstName: (value: string) => void;
  setLastName: (value: string) => void;
  setMiddleName: (value: string) => void;
  setPhone: (value: string) => void;
  validateEmailField: (value: string) => null | string;
  validateFirstName: (value: string) => null | string;
  validateLastName: (value: string) => null | string;
  validatePhoneField: (value: string) => null | string;
} => {
  const [firstName, setFirstName] = useState(profile.first_name);
  const [lastName, setLastName] = useState(profile.last_name);
  const [middleName, setMiddleName] = useState(profile.middle_name);
  const [email, setEmail] = useState(profile.email);
  const [phone, setPhoneValue] = useState(() => formatPhoneDisplay(profile.phone));

  const setPhone = (value: string): void => {
    setPhoneValue((previous) => formatPhoneInput(value, previous));
  };

  const handleSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const read = (key: string): string => {
      const value = formData.get(key);

      return typeof value === 'string' ? value.trim() : '';
    };

    const nextPhone = read('phone') || phone.trim();
    const nextEmail = read('email') || email.trim();
    const nextFirstName = read('firstName') || firstName.trim();
    const nextLastName = read('lastName') || lastName.trim();
    const nextMiddleName = read('middleName') || middleName.trim();

    setPhoneValue(formatPhoneDisplay(nextPhone));
    setEmail(nextEmail);
    setFirstName(nextFirstName);
    setLastName(nextLastName);
    setMiddleName(nextMiddleName);

    onSave({
      email: nextEmail,
      firstName: nextFirstName,
      lastName: nextLastName,
      middleName: nextMiddleName,
      phone: normalizePhone(nextPhone),
    });
  };

  return {
    email,
    firstName,
    handleSubmit,
    lastName,
    middleName,
    phone,
    setEmail,
    setFirstName,
    setLastName,
    setMiddleName,
    setPhone,
    validateEmailField: validateEmail,
    validateFirstName: validateRequired,
    validateLastName: validateRequired,
    validatePhoneField: validatePhone,
  };
};

type UsePasswordFormParams = {
  onSubmit: (payload: { newPassword: string; oldPassword: string }) => void;
};

export const usePasswordForm = ({
  onSubmit,
}: UsePasswordFormParams): {
  confirmPassword: string;
  handleSubmit: (event: FormEvent<HTMLFormElement>) => void;
  newPassword: string;
  oldPassword: string;
  reset: () => void;
  setConfirmPassword: (value: string) => void;
  setNewPassword: (value: string) => void;
  setOldPassword: (value: string) => void;
  validateConfirmPassword: (value: string) => null | string;
  validateNewPassword: (value: string) => null | string;
  validateOldPassword: (value: string) => null | string;
} => {
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const validateOldPassword = (value: string): null | string => validateRequired(value);

  const validateNewPassword = (value: string): null | string => {
    const requiredError = validateRequired(value);

    if (requiredError) {
      return requiredError;
    }

    if (value.length < 6) {
      return 'Минимум 6 символов';
    }

    return null;
  };

  const validateConfirmPassword = (value: string): null | string => {
    const requiredError = validateRequired(value);

    if (requiredError) {
      return requiredError;
    }

    if (value !== newPassword) {
      return 'Пароли не совпадают';
    }

    return null;
  };

  const handleSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    onSubmit({ newPassword, oldPassword });
  };

  const reset = (): void => {
    setOldPassword('');
    setNewPassword('');
    setConfirmPassword('');
  };

  return {
    confirmPassword,
    handleSubmit,
    newPassword,
    oldPassword,
    reset,
    setConfirmPassword,
    setNewPassword,
    setOldPassword,
    validateConfirmPassword,
    validateNewPassword,
    validateOldPassword,
  };
};

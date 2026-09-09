import {
  formatIncompletePhoneNumber,
  isValidPhoneNumber,
  parseIncompletePhoneNumber,
  parsePhoneNumberFromString,
} from 'libphonenumber-js';

export const EMAIL_INVALID_MESSAGE = 'Введите корректный email';
export const PHONE_INVALID_MESSAGE = 'Введите корректный номер телефона';
export const DEFAULT_PHONE_COUNTRY = 'RU' as const;

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type ValidateOptions = {
  required?: boolean;
};

export const isValidEmail = (value: string): boolean => {
  const email = value.trim();

  if (!email || email.includes('..')) {
    return false;
  }

  return EMAIL_PATTERN.test(email);
};

export const isValidPhone = (value: string): boolean => {
  const phone = String(value ?? '').trim();

  if (!phone) {
    return false;
  }

  return isValidPhoneNumber(phone, DEFAULT_PHONE_COUNTRY);
};

/** Нормализация в E.164 (`+79…`), иначе исходная строка. */
export const normalizePhone = (value: string): string => {
  const parsed = parsePhoneNumberFromString(String(value ?? '').trim(), DEFAULT_PHONE_COUNTRY);

  return parsed?.format('E.164') ?? String(value ?? '').trim();
};

/**
 * Форматирует номер по мере ввода (As You Type).
 * Учитывает backspace на скобках/пробелах — иначе значение «залипает».
 */
export const formatPhoneInput = (nextValue: string, previousValue = ''): string => {
  const previousDigits = parseIncompletePhoneNumber(previousValue);
  let nextDigits = parseIncompletePhoneNumber(nextValue);

  // Если стёрли только разделитель, цифры не меняются — убираем последнюю цифру вручную.
  if (nextDigits === previousDigits && nextValue.length < previousValue.length) {
    const formatted = formatIncompletePhoneNumber(nextDigits, DEFAULT_PHONE_COUNTRY);

    if (formatted.indexOf(nextValue) === 0) {
      nextDigits = nextDigits.slice(0, -1);
    }
  }

  return formatIncompletePhoneNumber(nextDigits, DEFAULT_PHONE_COUNTRY);
};

/** `null` — ок; иначе текст ошибки. */
export const validateEmail = (value: string, options: ValidateOptions = {}): null | string => {
  const required = options.required ?? true;
  const email = String(value ?? '').trim();

  if (!email) {
    return required ? 'Обязательное поле' : null;
  }

  return isValidEmail(email) ? null : EMAIL_INVALID_MESSAGE;
};

/** `null` — ок; иначе текст ошибки. */
export const validatePhone = (value: string, options: ValidateOptions = {}): null | string => {
  const required = options.required ?? true;
  const phone = String(value ?? '').trim();

  if (!phone) {
    return required ? 'Обязательное поле' : null;
  }

  return isValidPhone(phone) ? null : PHONE_INVALID_MESSAGE;
};

import {
  formatIncompletePhoneNumber,
  isValidPhoneNumber,
  parseIncompletePhoneNumber,
  parsePhoneNumberFromString,
} from 'libphonenumber-js';

export const REQUIRED_FIELD_MESSAGE = 'Обязательное поле';
export const EMAIL_INVALID_MESSAGE = 'Введите корректный email';
export const EMAIL_REQUIRED_MESSAGE = REQUIRED_FIELD_MESSAGE;
export const PHONE_INVALID_MESSAGE = 'Введите корректный номер телефона';
export const PHONE_REQUIRED_MESSAGE = REQUIRED_FIELD_MESSAGE;
export const INN_INVALID_MESSAGE = 'Введите корректный ИНН';
export const INN_LENGTH_MESSAGE = 'Введите ИНН из 10 или 12 цифр';
export const DEFAULT_PHONE_COUNTRY = 'RU' as const;

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const WEBSITE_PATTERN = /^(https?:\/\/)?([a-z0-9-]+\.)+[a-z]{2,}(\/.*)?$/i;

type ValidateOptions = {
  required?: boolean;
};

const onlyDigits = (value: string): string => String(value ?? '').replace(/\D/g, '');

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

/** Контрольная сумма ИНН (как на бэкенде auth). */
const isValidInnChecksum = (inn: string): boolean => {
  const digits = [...inn].map((char) => Number(char));

  if (digits.some((digit) => !Number.isInteger(digit))) {
    return false;
  }

  if (inn.length === 10) {
    const coeffs = [2, 4, 10, 3, 5, 9, 4, 6, 8];
    const sum = coeffs.reduce((acc, coeff, index) => acc + digits[index] * coeff, 0);

    return (sum % 11) % 10 === digits[9];
  }

  if (inn.length === 12) {
    const coeffs11 = [7, 2, 4, 10, 3, 5, 9, 4, 6, 8];
    const sum11 = coeffs11.reduce((acc, coeff, index) => acc + digits[index] * coeff, 0);
    const control11 = (sum11 % 11) % 10;

    if (control11 !== digits[10]) {
      return false;
    }

    const coeffs12 = [3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8];
    const sum12 = coeffs12.reduce((acc, coeff, index) => acc + digits[index] * coeff, 0);

    return (sum12 % 11) % 10 === digits[11];
  }

  return false;
};

/** Нормализация в E.164 (`+79…`), иначе исходная строка. */
export const normalizePhone = (value: string): string => {
  const parsed = parsePhoneNumberFromString(String(value ?? '').trim(), DEFAULT_PHONE_COUNTRY);

  return parsed?.format('E.164') ?? String(value ?? '').trim();
};

/**
 * Красивый вид для уже сохранённого / частичного номера (в т.ч. E.164).
 * @see https://github.com/catamphetamine/libphonenumber-js#as-you-type-formatter
 */
export const formatPhoneDisplay = (value: string): string => {
  const trimmed = String(value ?? '').trim();

  if (!trimmed) {
    return '';
  }

  return formatIncompletePhoneNumber(parseIncompletePhoneNumber(trimmed), DEFAULT_PHONE_COUNTRY);
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
export const validateRequired = (value: string): null | string => {
  return String(value ?? '').trim() ? null : REQUIRED_FIELD_MESSAGE;
};

/** `null` — ок; иначе текст ошибки. */
export const validateEmail = (value: string, options: ValidateOptions = {}): null | string => {
  const required = options.required ?? true;
  const email = String(value ?? '').trim();

  if (!email) {
    return required ? EMAIL_REQUIRED_MESSAGE : null;
  }

  return isValidEmail(email) ? null : EMAIL_INVALID_MESSAGE;
};

/** `null` — ок; иначе текст ошибки. */
export const validatePhone = (value: string, options: ValidateOptions = {}): null | string => {
  const required = options.required ?? true;
  const phone = String(value ?? '').trim();

  if (!phone) {
    return required ? PHONE_REQUIRED_MESSAGE : null;
  }

  return isValidPhone(phone) ? null : PHONE_INVALID_MESSAGE;
};

/** `null` — ок; иначе текст ошибки. ИНН юрлица — 10 цифр, ИП — 12 + checksum. */
export const validateInn = (value: string): null | string => {
  const inn = onlyDigits(value);

  if (!inn) {
    return REQUIRED_FIELD_MESSAGE;
  }

  if (inn.length !== 10 && inn.length !== 12) {
    return INN_LENGTH_MESSAGE;
  }

  return isValidInnChecksum(inn) ? null : INN_INVALID_MESSAGE;
};

/** Опциональное поле: пусто = ок, иначе строго `length` цифр. */
export const validateOptionalDigits = (
  value: string,
  length: number,
  message: string,
): null | string => {
  const digits = onlyDigits(value);

  if (!digits) {
    return null;
  }

  return digits.length === length ? null : message;
};

/** Опциональное поле: пусто = ок, иначе одна из допустимых длин. */
export const validateOptionalDigitLengths = (
  value: string,
  lengths: number[],
  message: string,
): null | string => {
  const digits = onlyDigits(value);

  if (!digits) {
    return null;
  }

  return lengths.includes(digits.length) ? null : message;
};

/** Опциональный сайт. */
export const validateOptionalWebsite = (value: string): null | string => {
  const website = String(value ?? '').trim();

  if (!website) {
    return null;
  }

  return WEBSITE_PATTERN.test(website) ? null : 'Введите корректный адрес сайта';
};

/** Пароль: обязательный, минимум 6 символов. */
export const validatePassword = (value: string): null | string => {
  const password = String(value ?? '');

  if (!password.trim()) {
    return REQUIRED_FIELD_MESSAGE;
  }

  if (password.length < 6) {
    return 'Пароль должен быть не короче 6 символов';
  }

  return null;
};

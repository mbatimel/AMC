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
export const DEFAULT_PHONE_COUNTRY = 'RU' as const;

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

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

type ValidateOptions = {
  required?: boolean;
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

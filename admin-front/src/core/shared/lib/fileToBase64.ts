export const DOCUMENT_FILE_ACCEPT = '.pdf,.jpg,.jpeg,.png,.webp';
export const DOCUMENT_FILE_HINT = 'PDF, JPG, PNG или WEBP · до 8 МБ';
export const IMAGE_FILE_ACCEPT = '.jpg,.jpeg,.png,.webp';
export const IMAGE_FILE_HINT = 'JPG, PNG или WEBP · до 8 МБ';
/** Декодированный файл: бэкенд до 10 МиБ, но JSON+base64 не влезает в BodyLimit 11 МиБ. */
export const MAX_DOCUMENT_FILE_BYTES = 8 * 1024 * 1024;

export const isPdfFileUrl = (url: string): boolean => /\.pdf(?:$|[?#])/i.test(url);

export const fileToBase64 = (file: File): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = () => {
      const result = reader.result;

      if (typeof result !== 'string') {
        reject(new Error('Не удалось прочитать файл'));
        return;
      }

      const commaIndex = result.indexOf(',');

      resolve(commaIndex >= 0 ? result.slice(commaIndex + 1) : result);
    };
    reader.onerror = () => reject(new Error('Не удалось прочитать файл'));
    reader.readAsDataURL(file);
  });

export const assertDocumentFile = (file: File): void => {
  if (file.size > MAX_DOCUMENT_FILE_BYTES) {
    throw new Error('Файл больше 8 МБ. Выберите файл меньшего размера');
  }
};

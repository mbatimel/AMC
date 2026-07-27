export type SplitFullNameResult = {
  name: string;
  surename: string;
};

/** ФИО: «Фамилия Имя Отчество» → surename / name (как ждёт auth API). */
export const splitFullName = (fullName: string): SplitFullNameResult => {
  const parts = fullName.trim().split(/\s+/).filter(Boolean);

  if (parts.length === 0) {
    return { name: '', surename: '' };
  }

  if (parts.length === 1) {
    return { name: parts[0], surename: parts[0] };
  }

  return {
    name: parts[1],
    surename: parts[0],
  };
};

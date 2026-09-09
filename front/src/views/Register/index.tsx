'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

import {
  $authError,
  $isAuthPending,
  buildRegisterPayload,
  signupFx,
} from '@/core/entities/session';
import { IconUserPlus } from '@/core/shared/icons/IconUserPlus';
import {
  formatPhoneInput,
  normalizePhone,
  validateEmail,
  validateInn,
  validateOptionalDigitLengths,
  validateOptionalDigits,
  validateOptionalWebsite,
  validatePassword,
  validatePhone,
  validateRequired,
} from '@/core/shared/lib/validateContact';
import { AppPath } from '@/core/shared/router/paths';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

/**
 * Валидация через TextField.validate (React Aria):
 * — ошибки и красная рамка на поле;
 * — фокус на первое невалидное;
 * — submit не уходит, пока форма невалидна.
 *
 * Обязательные по ТЗ (из того, что уже есть в форме):
 * email, password, контактное лицо, телефон, ИНН, название компании.
 * Город и файл реквизитов в форме пока нет — не добавляем.
 * Остальные поля опциональны; если заполнены — проверяем формат.
 */
export const Register = (): JSX.Element => {
  const router = useRouter();
  const [inn, setInn] = useState('');
  const [phone, setPhoneValue] = useState('');
  const [phoneAdditional, setPhoneAdditionalValue] = useState('');
  const [authError, isPending, signup] = useUnit([$authError, $isAuthPending, signupFx]);

  const setPhone = (value: string): void => {
    setPhoneValue((previous) => formatPhoneInput(value, previous));
  };

  const setPhoneAdditional = (value: string): void => {
    setPhoneAdditionalValue((previous) => formatPhoneInput(value, previous));
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);

    if (phone.trim()) {
      formData.set('phone', normalizePhone(phone));
    } else {
      formData.delete('phone');
    }

    if (phoneAdditional.trim()) {
      formData.set('phoneAdditional', normalizePhone(phoneAdditional));
    } else {
      formData.delete('phoneAdditional');
    }

    formData.set('inn', inn.replace(/\D/g, ''));

    void signup(buildRegisterPayload(formData))
      .then(() => {
        router.replace(AppPath.Login);
      })
      .catch(() => {
        // error in $authError
      });
  };

  return (
    <AuthShell wide>
      <AuthCardHeader
        description="Заполните карточку организации или ИП. После регистрации войдите в личный кабинет."
        icon={IconUserPlus}
        title="Регистрация клиента"
      />

      <Form className={clsx(formStyles.form)} onSubmit={handleSubmit}>
        <section className={clsx(formStyles.section)}>
          <h2 className={clsx(formStyles.sectionTitle)}>Организация</h2>
          <div className={clsx(formStyles.grid)}>
            <TextField
              className={clsx(formStyles.field, formStyles.gridFull)}
              name="fullCompanyName"
            >
              <Label className={clsx(formStyles.label)}>Полное наименование</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                placeholder="Общество с ограниченной ответственностью «…»"
              />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              isRequired
              name="shortCompanyName"
              validate={validateRequired}
            >
              <Label className={clsx(formStyles.label)}>Краткое наименование</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="ООО «…»" />
              <FieldError />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              isRequired
              name="inn"
              onChange={(value) => setInn(value.replace(/\D/g, '').slice(0, 12))}
              validate={validateInn}
              value={inn}
            >
              <Label className={clsx(formStyles.label)}>ИНН</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                inputMode="numeric"
                placeholder="10 или 12 цифр"
              />
              <FieldError />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              name="kpp"
              validate={(value) => validateOptionalDigits(value, 9, 'КПП должен содержать 9 цифр')}
            >
              <Label className={clsx(formStyles.label)}>КПП</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="9 цифр" />
              <FieldError />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              name="ogrn"
              validate={(value) =>
                validateOptionalDigitLengths(value, [13, 15], 'ОГРН — 13 цифр, ОГРНИП — 15')
              }
            >
              <Label className={clsx(formStyles.label)}>ОГРН / ОГРНИП</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="13 или 15 цифр" />
              <FieldError />
            </TextField>
            <TextField className={clsx(formStyles.field)} name="okved">
              <Label className={clsx(formStyles.label)}>Основной ОКВЭД</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="напр. 46.74" />
            </TextField>
            <TextField className={clsx(formStyles.field)} name="taxSystem">
              <Label className={clsx(formStyles.label)}>Система налогообложения</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="ОСНО (с НДС)" />
            </TextField>
          </div>
        </section>

        <section className={clsx(formStyles.section)}>
          <h2 className={clsx(formStyles.sectionTitle)}>Место нахождения</h2>
          <div className={clsx(formStyles.grid)}>
            <TextField className={clsx(formStyles.field, formStyles.gridFull)} name="legalAddress">
              <Label className={clsx(formStyles.label)}>Юридический адрес</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                placeholder="Индекс, регион, город, улица, дом"
              />
            </TextField>
            <TextField className={clsx(formStyles.field, formStyles.gridFull)} name="actualAddress">
              <Label className={clsx(formStyles.label)}>Фактический адрес</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                placeholder="Совпадает с юридическим — оставьте пустым"
              />
            </TextField>
          </div>
        </section>

        <section className={clsx(formStyles.section)}>
          <h2 className={clsx(formStyles.sectionTitle)}>Руководитель</h2>
          <div className={clsx(formStyles.grid)}>
            <TextField
              className={clsx(formStyles.field)}
              isRequired
              name="directorFullName"
              validate={validateRequired}
            >
              <Label className={clsx(formStyles.label)}>Контактное лицо</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                placeholder="Иванов Иван Иванович"
              />
              <FieldError />
            </TextField>
            <TextField className={clsx(formStyles.field)} name="directorPosition">
              <Label className={clsx(formStyles.label)}>Должность</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                placeholder="Генеральный директор"
              />
            </TextField>
          </div>
        </section>

        <section className={clsx(formStyles.section)}>
          <h2 className={clsx(formStyles.sectionTitle)}>Контакты</h2>
          <div className={clsx(formStyles.grid)}>
            <TextField
              className={clsx(formStyles.field)}
              isRequired
              name="phone"
              onChange={setPhone}
              validate={(value) => validatePhone(value, { required: true })}
              value={phone}
            >
              <Label className={clsx(formStyles.label)}>Контактный телефон</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                placeholder="+7 999 123 45 67"
                type="tel"
              />
              <FieldError />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              name="phoneAdditional"
              onChange={setPhoneAdditional}
              validate={(value) => validatePhone(value, { required: false })}
              value={phoneAdditional}
            >
              <Label className={clsx(formStyles.label)}>Дополнительный телефон</Label>
              <Input
                className={clsx(formStyles.input)}
                fullWidth
                placeholder="+7 999 123 45 67"
                type="tel"
              />
              <FieldError />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              isRequired
              name="email"
              type="email"
              validate={validateEmail}
            >
              <Label className={clsx(formStyles.label)}>E-mail</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="info@company.ru" />
              <FieldError />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              name="website"
              validate={validateOptionalWebsite}
            >
              <Label className={clsx(formStyles.label)}>Сайт</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="https://" />
              <FieldError />
            </TextField>
          </div>
        </section>

        <section className={clsx(formStyles.section)}>
          <h2 className={clsx(formStyles.sectionTitle)}>Банковские реквизиты</h2>
          <div className={clsx(formStyles.grid)}>
            <TextField
              className={clsx(formStyles.field)}
              name="bankAccount"
              validate={(value) =>
                validateOptionalDigits(value, 20, 'Расчётный счёт должен содержать 20 цифр')
              }
            >
              <Label className={clsx(formStyles.label)}>Расчётный счёт</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="20 цифр" />
              <FieldError />
            </TextField>
            <TextField className={clsx(formStyles.field)} name="bankName">
              <Label className={clsx(formStyles.label)}>Наименование банка</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="ПАО «…»" />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              name="bik"
              validate={(value) => validateOptionalDigits(value, 9, 'БИК должен содержать 9 цифр')}
            >
              <Label className={clsx(formStyles.label)}>БИК</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="9 цифр" />
              <FieldError />
            </TextField>
            <TextField
              className={clsx(formStyles.field)}
              name="corrAccount"
              validate={(value) =>
                validateOptionalDigits(value, 20, 'Корреспондентский счёт должен содержать 20 цифр')
              }
            >
              <Label className={clsx(formStyles.label)}>Корреспондентский счёт</Label>
              <Input className={clsx(formStyles.input)} fullWidth placeholder="20 цифр" />
              <FieldError />
            </TextField>
          </div>
        </section>

        <TextField
          className={clsx(formStyles.field)}
          isRequired
          name="password"
          type="password"
          validate={validatePassword}
        >
          <Label className={clsx(formStyles.label)}>Пароль</Label>
          <Input className={clsx(formStyles.input)} fullWidth placeholder="••••••••" />
          <FieldError />
        </TextField>

        {authError ? <p className={clsx(formStyles.error)}>{authError}</p> : null}

        <div className={clsx(formStyles.actions)}>
          <Button
            className={clsx(formStyles.submitButton)}
            isDisabled={isPending}
            type="submit"
            variant="primary"
          >
            <IconUserPlus currentColor="currentColor" height={16} width={16} />
            {isPending ? 'Регистрация…' : 'Зарегистрироваться'}
          </Button>
          <Link className={clsx(formStyles.secondaryLink)} href={AppPath.Login}>
            ← Уже есть аккаунт
          </Link>
        </div>
        <p className={clsx(formStyles.hint)}>
          Нажимая «Зарегистрироваться», вы соглашаетесь на обработку персональных данных.
        </p>
      </Form>
    </AuthShell>
  );
};

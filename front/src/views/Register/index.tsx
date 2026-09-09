'use client';

import { Button, FieldError, Form, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useId, useState } from 'react';

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
  REQUISITES_FILE_ACCEPT,
  validateEmail,
  validateInn,
  validatePassword,
  validatePhone,
  validateRequired,
  validateRequisitesFile,
} from '@/core/shared/lib/validateContact';
import { AppPath } from '@/core/shared/router/paths';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

/**
 * Обязательные поля: email, password, контактное лицо, телефон, ИНН,
 * краткое наименование, файл с реквизитами.
 */
export const Register = (): JSX.Element => {
  const router = useRouter();
  const fileInputId = useId();
  const [inn, setInn] = useState('');
  const [phone, setPhoneValue] = useState('');
  const [requisitesFile, setRequisitesFile] = useState<File | null>(null);
  const [requisitesError, setRequisitesError] = useState<null | string>(null);
  const [authError, isPending, signup] = useUnit([$authError, $isAuthPending, signupFx]);

  const setPhone = (value: string): void => {
    setPhoneValue((previous) => formatPhoneInput(value, previous));
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
    const nextFile = event.target.files?.[0] ?? null;

    setRequisitesFile(nextFile);
    setRequisitesError(validateRequisitesFile(nextFile));
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const fileError = validateRequisitesFile(requisitesFile);

    setRequisitesError(fileError);

    if (fileError || !requisitesFile) {
      return;
    }

    const formData = new FormData(event.currentTarget);

    formData.set('phone', normalizePhone(phone));
    formData.set('inn', inn.replace(/\D/g, ''));
    formData.set('requisitesFile', requisitesFile);

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
        description="Укажите данные организации — после регистрации войдите в личный кабинет."
        icon={IconUserPlus}
        title="Регистрация клиента"
      />

      <Form className={clsx(formStyles.form)} onSubmit={handleSubmit}>
        <div className={clsx(formStyles.grid)}>
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
          <TextField
            className={clsx(formStyles.field)}
            isRequired
            name="phone"
            onChange={setPhone}
            validate={(value) => validatePhone(value, { required: true })}
            value={phone}
          >
            <Label className={clsx(formStyles.label)}>Телефон</Label>
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
            isRequired
            name="password"
            type="password"
            validate={validatePassword}
          >
            <Label className={clsx(formStyles.label)}>Пароль</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="••••••••" />
            <FieldError />
          </TextField>
          <div
            className={clsx(
              formStyles.fileField,
              formStyles.gridFull,
              requisitesError && formStyles.fileFieldInvalid,
            )}
          >
            <label className={clsx(formStyles.label)} htmlFor={fileInputId}>
              Файл с реквизитами
              <span aria-hidden className={clsx(formStyles.requiredMark)}>
                *
              </span>
            </label>
            <label className={clsx(formStyles.fileDrop)} htmlFor={fileInputId}>
              <span className={clsx(formStyles.fileDropTitle)}>
                {requisitesFile ? requisitesFile.name : 'Выберите или перетащите файл'}
              </span>
              <span className={clsx(formStyles.fileDropHint)}>
                PDF, DOC, DOCX, JPG или PNG · до 10 МБ
              </span>
              <input
                accept={REQUISITES_FILE_ACCEPT}
                className={clsx(formStyles.fileInput)}
                id={fileInputId}
                name="requisitesFile"
                onChange={handleFileChange}
                required
                type="file"
              />
            </label>
            {requisitesError ? (
              <p className={clsx(formStyles.fileError)} role="alert">
                {requisitesError}
              </p>
            ) : null}
          </div>
        </div>

        {authError ? <p className={clsx(formStyles.error)}>{authError}</p> : null}

        <div className={clsx(formStyles.actions)}>
          <Button
            className={clsx(formStyles.submitButton, formStyles.submitButtonBlock)}
            isDisabled={isPending}
            type="submit"
            variant="primary"
          >
            <IconUserPlus currentColor="currentColor" height={16} width={16} />
            {isPending ? 'Регистрация…' : 'Зарегистрироваться'}
          </Button>
        </div>

        <p className={clsx(formStyles.hint)}>
          Нажимая «Зарегистрироваться», вы соглашаетесь на обработку персональных данных.
        </p>
      </Form>

      <div className={clsx(formStyles.footer)}>
        <span>Уже есть аккаунт?</span>
        <Link className={clsx(formStyles.footerLink)} href={AppPath.Login}>
          Войти
        </Link>
      </div>
    </AuthShell>
  );
};

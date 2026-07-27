'use client';

import { Button, FieldError, Form, Input, Label, Tab, TabList, Tabs, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

import { $authError, $isAuthPending, signupFx, splitFullName } from '@/core/entities/session';
import { IconBuilding } from '@/core/shared/icons/IconBuilding';
import { IconUser } from '@/core/shared/icons/IconUser';
import { IconUserPlus } from '@/core/shared/icons/IconUserPlus';
import { readFormString } from '@/core/shared/lib/readFormString';
import { AppPath } from '@/core/shared/router/paths';
import { AuthShell } from '@/core/shared/ui/AuthShell';
import { AuthCardHeader } from '@/core/shared/ui/AuthShell/AuthCardHeader';
import formStyles from '@/core/shared/ui/AuthShell/AuthForm.module.css';

import styles from './Register.module.css';

enum RegisterType {
  Individual = 'individual',
  Organization = 'organization',
}

export const Register = (): JSX.Element => {
  const router = useRouter();
  const [registerType, setRegisterType] = useState<RegisterType>(RegisterType.Organization);
  const [authError, isPending, signup] = useUnit([$authError, $isAuthPending, signupFx]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const email = readFormString(formData, 'email').trim();
    const password = readFormString(formData, 'password');
    const fullNameKey =
      registerType === RegisterType.Organization ? 'directorFullName' : 'fullName';
    const fullName = readFormString(formData, fullNameKey).trim();
    const { name, surename } = splitFullName(fullName);

    void signup({ email, name, password, surename })
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
        description="Заполните карточку клиента. После регистрации войдите в личный кабинет."
        icon={IconUserPlus}
        title="Регистрация клиента"
      />

      <Tabs
        className={clsx(formStyles.tabs, styles.tabs)}
        onSelectionChange={(key) => {
          setRegisterType(String(key) as RegisterType);
        }}
        selectedKey={registerType}
      >
        <TabList aria-label="Тип клиента" className={clsx(styles.tabList)}>
          <Tab className={clsx(styles.tab)} id={RegisterType.Organization}>
            <IconBuilding height={14} width={14} />
            Организация / ИП
          </Tab>
          <Tab className={clsx(styles.tab)} id={RegisterType.Individual}>
            <IconUser height={14} width={14} />
            Физическое лицо
          </Tab>
        </TabList>
      </Tabs>

      <Form className={clsx(formStyles.form)} onSubmit={handleSubmit}>
        {registerType === RegisterType.Organization ? <OrganizationFields /> : <IndividualFields />}

        <TextField className={clsx(formStyles.field)} isRequired name="password" type="password">
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

const OrganizationFields = (): JSX.Element => {
  return (
    <>
      <section className={clsx(formStyles.section)}>
        <h2 className={clsx(formStyles.sectionTitle)}>Организация</h2>
        <div className={clsx(formStyles.grid)}>
          <TextField className={clsx(formStyles.field, formStyles.gridFull)} name="fullCompanyName">
            <Label className={clsx(formStyles.label)}>Полное наименование</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Общество с ограниченной ответственностью «…»" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="shortCompanyName">
            <Label className={clsx(formStyles.label)}>Краткое наименование</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="ООО «…»" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="inn">
            <Label className={clsx(formStyles.label)}>ИНН</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="10 или 12 цифр" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="kpp">
            <Label className={clsx(formStyles.label)}>КПП</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="9 цифр" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="ogrn">
            <Label className={clsx(formStyles.label)}>ОГРН / ОГРНИП</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="13 или 15 цифр" />
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
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Индекс, регион, город, улица, дом" />
          </TextField>
          <TextField className={clsx(formStyles.field, formStyles.gridFull)} name="actualAddress">
            <Label className={clsx(formStyles.label)}>Фактический адрес</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Совпадает с юридическим — оставьте пустым" />
          </TextField>
        </div>
      </section>

      <section className={clsx(formStyles.section)}>
        <h2 className={clsx(formStyles.sectionTitle)}>Руководитель</h2>
        <div className={clsx(formStyles.grid)}>
          <TextField className={clsx(formStyles.field)} isRequired name="directorFullName">
            <Label className={clsx(formStyles.label)}>ФИО руководителя</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Иванов Иван Иванович" />
            <FieldError />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="directorPosition">
            <Label className={clsx(formStyles.label)}>Должность</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Генеральный директор" />
          </TextField>
        </div>
      </section>

      <section className={clsx(formStyles.section)}>
        <h2 className={clsx(formStyles.sectionTitle)}>Контакты</h2>
        <div className={clsx(formStyles.grid)}>
          <TextField className={clsx(formStyles.field)} name="phone">
            <Label className={clsx(formStyles.label)}>Контактный телефон</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="+7" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="phoneAdditional">
            <Label className={clsx(formStyles.label)}>Дополнительный телефон</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="+7" />
          </TextField>
          <TextField className={clsx(formStyles.field)} isRequired name="email" type="email">
            <Label className={clsx(formStyles.label)}>E-mail</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="info@company.ru" />
            <FieldError />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="website">
            <Label className={clsx(formStyles.label)}>Сайт</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="https://" />
          </TextField>
        </div>
      </section>

      <section className={clsx(formStyles.section)}>
        <h2 className={clsx(formStyles.sectionTitle)}>Банковские реквизиты</h2>
        <div className={clsx(formStyles.grid)}>
          <TextField className={clsx(formStyles.field)} name="bankAccount">
            <Label className={clsx(formStyles.label)}>Расчётный счёт</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="20 цифр" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="bankName">
            <Label className={clsx(formStyles.label)}>Наименование банка</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="ПАО «…»" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="bik">
            <Label className={clsx(formStyles.label)}>БИК</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="9 цифр" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="corrAccount">
            <Label className={clsx(formStyles.label)}>Корреспондентский счёт</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="20 цифр" />
          </TextField>
        </div>
      </section>
    </>
  );
};

const IndividualFields = (): JSX.Element => {
  return (
    <>
      <section className={clsx(formStyles.section)}>
        <h2 className={clsx(formStyles.sectionTitle)}>Личные данные</h2>
        <div className={clsx(formStyles.grid)}>
          <TextField className={clsx(formStyles.field, formStyles.gridFull)} isRequired name="fullName">
            <Label className={clsx(formStyles.label)}>ФИО</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Иванов Иван Иванович" />
            <FieldError />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="inn">
            <Label className={clsx(formStyles.label)}>ИНН (необязательно)</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="12 цифр" />
          </TextField>
          <TextField className={clsx(formStyles.field)} name="city">
            <Label className={clsx(formStyles.label)}>Город</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Самара" />
          </TextField>
        </div>
      </section>

      <section className={clsx(formStyles.section)}>
        <h2 className={clsx(formStyles.sectionTitle)}>Контакты и доставка</h2>
        <div className={clsx(formStyles.grid)}>
          <TextField className={clsx(formStyles.field)} name="phone">
            <Label className={clsx(formStyles.label)}>Телефон</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="+7" />
          </TextField>
          <TextField className={clsx(formStyles.field)} isRequired name="email" type="email">
            <Label className={clsx(formStyles.label)}>E-mail</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="ivan@mail.ru" />
            <FieldError />
          </TextField>
          <TextField className={clsx(formStyles.field, formStyles.gridFull)} name="deliveryAddress">
            <Label className={clsx(formStyles.label)}>Адрес доставки</Label>
            <Input className={clsx(formStyles.input)} fullWidth placeholder="Индекс, город, улица, дом, квартира" />
          </TextField>
        </div>
      </section>
    </>
  );
};

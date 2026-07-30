'use client';

import {
  Alert,
  Button,
  Card,
  Chip,
  Description,
  EmptyState,
  FieldError,
  Form,
  Input,
  Label,
  Spinner,
  Surface,
  TextField,
  Typography,
} from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useState } from 'react';

import type { Profile } from '@/core/shared/api/profile';

import { formatPrice } from '@/core/shared/lib/formatPrice';

import styles from '../Cabinet.module.css';
import {
  $clientDetails,
  $clients,
  $conditions,
  $isProfilePending,
  $profile,
  $profileError,
  clientActivateRequested,
  passwordChangeRequested,
  profileSaveRequested,
} from '../model/profile';

type ProfileUserFormProps = {
  onSave: (payload: {
    email: string;
    firstName: string;
    lastName: string;
    middleName: string;
    phone: string;
  }) => void;
  pending: boolean;
  profile: Profile;
};

const ProfileUserForm = ({ onSave, pending, profile }: ProfileUserFormProps): JSX.Element => {
  const [firstName, setFirstName] = useState(profile.first_name);
  const [lastName, setLastName] = useState(profile.last_name);
  const [middleName, setMiddleName] = useState(profile.middle_name);
  const [email, setEmail] = useState(profile.email);
  const [phone, setPhone] = useState(profile.phone);

  return (
    <Form
      onSubmit={(event) => {
        event.preventDefault();
        onSave({
          email: email.trim(),
          firstName: firstName.trim(),
          lastName: lastName.trim(),
          middleName: middleName.trim(),
          phone: phone.trim(),
        });
      }}
    >
      <div className={clsx(styles.formGrid)}>
        <TextField isRequired name="lastName" onChange={setLastName} value={lastName}>
          <Label>Фамилия</Label>
          <Input />
        </TextField>
        <TextField isRequired name="firstName" onChange={setFirstName} value={firstName}>
          <Label>Имя</Label>
          <Input />
        </TextField>
        <TextField name="middleName" onChange={setMiddleName} value={middleName}>
          <Label>Отчество</Label>
          <Input />
        </TextField>
        <TextField isRequired name="phone" onChange={setPhone} value={phone}>
          <Label>Телефон</Label>
          <Input type="tel" />
        </TextField>
        <TextField isRequired name="email" onChange={setEmail} value={email}>
          <Label>Email</Label>
          <Input type="email" />
        </TextField>
      </div>
      <div className={clsx(styles.formActions)}>
        <Button isDisabled={pending} type="submit" variant="primary">
          Сохранить
        </Button>
      </div>
    </Form>
  );
};

const PasswordForm = ({
  onSubmit,
  pending,
}: {
  onSubmit: (payload: { newPassword: string; oldPassword: string }) => void;
  pending: boolean;
}): JSX.Element => {
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState<null | string>(null);

  return (
    <Form
      onSubmit={(event) => {
        event.preventDefault();

        if (newPassword !== confirmPassword) {
          setPasswordError('Пароли не совпадают');

          return;
        }

        if (newPassword.length < 6) {
          setPasswordError('Минимум 6 символов');

          return;
        }

        setPasswordError(null);
        onSubmit({ newPassword, oldPassword });
        setOldPassword('');
        setNewPassword('');
        setConfirmPassword('');
      }}
    >
      <div className={clsx(styles.formGrid)}>
        <TextField isRequired name="oldPassword" onChange={setOldPassword} value={oldPassword}>
          <Label>Текущий пароль</Label>
          <Input type="password" />
        </TextField>
        <TextField isRequired name="newPassword" onChange={setNewPassword} value={newPassword}>
          <Label>Новый пароль</Label>
          <Input type="password" />
        </TextField>
        <TextField
          isRequired
          name="confirmPassword"
          onChange={setConfirmPassword}
          value={confirmPassword}
        >
          <Label>Повтор нового пароля</Label>
          <Input type="password" />
          {passwordError ? <FieldError>{passwordError}</FieldError> : null}
        </TextField>
      </div>
      <div className={clsx(styles.formActions)}>
        <Button isDisabled={pending} type="submit" variant="secondary">
          Изменить пароль
        </Button>
      </div>
    </Form>
  );
};

type ConditionCardProps = {
  label: string;
  value: string;
};

const ConditionCard = ({ label, value }: ConditionCardProps): JSX.Element => {
  return (
    <Card>
      <Card.Header>
        <Card.Description>{label}</Card.Description>
        <Card.Title>{value}</Card.Title>
      </Card.Header>
    </Card>
  );
};

export const CabinetProfile = (): JSX.Element => {
  const [
    profile,
    clients,
    details,
    conditions,
    pending,
    error,
    saveProfile,
    activateClient,
    changePassword,
  ] = useUnit([
    $profile,
    $clients,
    $clientDetails,
    $conditions,
    $isProfilePending,
    $profileError,
    profileSaveRequested,
    clientActivateRequested,
    passwordChangeRequested,
  ]);

  const client = details ?? profile?.active_client;

  return (
    <div className={clsx(styles.main)}>
      <div className={clsx(styles.pageHeader)}>
        <div>
          <Typography.Heading className={clsx(styles.pageTitle)} level={1}>
            Профиль и условия
          </Typography.Heading>
          <Description className={clsx(styles.pageSubtitle)}>
            Данные пользователя, реквизиты и условия работы
          </Description>
        </div>
      </div>

      {error ? (
        <Alert className={clsx(styles.alert)} status="danger">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Content>
        </Alert>
      ) : null}

      {pending && !profile ? (
        <div className={clsx(styles.statusRow)}>
          <Spinner size="sm" />
          <Description>Загрузка…</Description>
        </div>
      ) : null}

      {profile ? (
        <Surface className={clsx(styles.panel, styles.profilePanel)}>
          <section className={clsx(styles.section)}>
            <Typography.Heading className={clsx(styles.sectionTitle)} level={3}>
              Данные пользователя
            </Typography.Heading>
            <ProfileUserForm
              key={profile.user_id}
              onSave={saveProfile}
              pending={pending}
              profile={profile}
            />
          </section>

          <section className={clsx(styles.section)}>
            <Typography.Heading className={clsx(styles.sectionTitle)} level={3}>
              Кабинеты / клиенты
            </Typography.Heading>
            {clients.length === 0 ? (
              <EmptyState>
                <Description>Нет привязанных кабинетов</Description>
              </EmptyState>
            ) : (
              <div className={clsx(styles.clientsList)}>
                {clients.map((item) => (
                  <Card key={item.client.id}>
                    <Card.Content className={clsx(styles.clientRow)}>
                      <div>
                        <div className={clsx(styles.clientName)}>
                          <Card.Title>
                            {item.client.company_name || item.client.contact_name || 'Клиент'}
                          </Card.Title>
                          {item.is_active ? (
                            <Chip className={clsx(styles.badge)} color="accent" size="sm">
                              <Chip.Label>Активный</Chip.Label>
                            </Chip>
                          ) : null}
                        </div>
                        <Card.Description>
                          {[item.client.inn && `ИНН ${item.client.inn}`, item.client.company_type]
                            .filter(Boolean)
                            .join(' · ') || item.client.id}
                        </Card.Description>
                      </div>
                      {!item.is_active ? (
                        <Button
                          isDisabled={pending}
                          onPress={() => activateClient(item.client.id)}
                          size="sm"
                          variant="secondary"
                        >
                          Сделать активным
                        </Button>
                      ) : null}
                    </Card.Content>
                  </Card>
                ))}
              </div>
            )}
          </section>

          <section className={clsx(styles.section)}>
            <Typography.Heading className={clsx(styles.sectionTitle)} level={3}>
              Реквизиты
            </Typography.Heading>
            {client ? (
              <dl className={clsx(styles.requisites)}>
                <dt>Название</dt>
                <dd>{client.company_name || '—'}</dd>
                <dt>Тип</dt>
                <dd>{client.company_type || '—'}</dd>
                <dt>ИНН</dt>
                <dd>{client.inn || '—'}</dd>
                <dt>ОГРН</dt>
                <dd>{client.ogrn || '—'}</dd>
                <dt>Адрес</dt>
                <dd>{client.address || '—'}</dd>
                <dt>Контакт</dt>
                <dd>{client.contact_name || '—'}</dd>
                <dt>Телефон</dt>
                <dd>{client.phone || '—'}</dd>
                <dt>Email</dt>
                <dd>{client.email || '—'}</dd>
              </dl>
            ) : (
              <EmptyState>
                <Description>Реквизиты недоступны</Description>
              </EmptyState>
            )}
          </section>

          <section className={clsx(styles.section)}>
            <Typography.Heading className={clsx(styles.sectionTitle)} level={3}>
              Условия работы
            </Typography.Heading>
            {conditions ? (
              <>
                <div className={clsx(styles.conditionsGrid)}>
                  <ConditionCard label="Ценовая группа" value={conditions.price_group || '—'} />
                  <ConditionCard
                    label="Кредитный лимит"
                    value={formatPrice(conditions.credit_limit)}
                  />
                  <ConditionCard label="Использовано" value={formatPrice(conditions.credit_used)} />
                  <ConditionCard label="Условия оплаты" value={conditions.payment_terms || '—'} />
                  <ConditionCard label="Менеджер" value={conditions.sales_contact || '—'} />
                  <ConditionCard label="Канал связи" value={conditions.contact_channel || '—'} />
                </div>

                {conditions.category_discounts.length > 0 ? (
                  <>
                    <Typography.Heading className={clsx(styles.sectionTitle)} level={4}>
                      Скидки по категориям
                    </Typography.Heading>
                    <div className={clsx(styles.discounts)}>
                      {conditions.category_discounts.map((discount) => (
                        <Chip
                          key={`${discount.category_id}-${discount.discount_percent}`}
                          variant="soft"
                        >
                          <Chip.Label>
                            {discount.category_name || discount.category_id}:{' '}
                            {discount.discount_percent}%
                          </Chip.Label>
                        </Chip>
                      ))}
                    </div>
                  </>
                ) : null}
              </>
            ) : (
              <EmptyState>
                <Description>Условия пока недоступны</Description>
              </EmptyState>
            )}
          </section>

          <section className={clsx(styles.section)}>
            <Typography.Heading className={clsx(styles.sectionTitle)} level={3}>
              Смена пароля
            </Typography.Heading>
            <PasswordForm onSubmit={changePassword} pending={pending} />
          </section>
        </Surface>
      ) : null}
    </div>
  );
};

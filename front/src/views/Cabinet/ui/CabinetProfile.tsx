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
import { useEffect } from 'react';

import type { Profile } from '@/core/shared/api/profile';

import { formatPrice } from '@/core/shared/lib/formatPrice';

import styles from '../Cabinet.module.css';
import { type ProfileFormSavePayload, usePasswordForm, useProfileForm } from '../lib/useProfile';
import {
  $clientDetails,
  $clients,
  $conditions,
  $isProfilePending,
  $passwordChangeVersion,
  $profile,
  $profileError,
  changePasswordFx,
  clientActivateRequested,
  passwordChangeRequested,
  profileSaveRequested,
} from '../model/profile';

type ProfileUserFormProps = {
  onSave: (payload: ProfileFormSavePayload) => void;
  pending: boolean;
  profile: Profile;
};

const ProfileUserForm = ({ onSave, pending, profile }: ProfileUserFormProps): JSX.Element => {
  const form = useProfileForm({ onSave, profile });

  return (
    <Form onSubmit={form.handleSubmit}>
      <div className={clsx(styles.formGrid)}>
        <TextField
          className={clsx(styles.formField)}
          isRequired
          name="firstName"
          onChange={form.setFirstName}
          validate={form.validateFirstName}
          value={form.firstName}
        >
          <Label className={clsx(styles.formLabel)}>Имя</Label>
          <Input className={clsx(styles.formInput)} fullWidth placeholder="Иван" />
          <FieldError />
        </TextField>
        <TextField
          className={clsx(styles.formField)}
          isRequired
          name="lastName"
          onChange={form.setLastName}
          validate={form.validateLastName}
          value={form.lastName}
        >
          <Label className={clsx(styles.formLabel)}>Фамилия</Label>
          <Input className={clsx(styles.formInput)} fullWidth placeholder="Иванов" />
          <FieldError />
        </TextField>
        <TextField
          className={clsx(styles.formField)}
          name="middleName"
          onChange={form.setMiddleName}
          value={form.middleName}
        >
          <Label className={clsx(styles.formLabel)}>Отчество</Label>
          <Input className={clsx(styles.formInput)} fullWidth placeholder="Иванович" />
        </TextField>
        <TextField
          className={clsx(styles.formField)}
          isRequired
          name="phone"
          onChange={form.setPhone}
          validate={form.validatePhoneField}
          value={form.phone}
        >
          <Label className={clsx(styles.formLabel)}>Телефон</Label>
          <Input
            className={clsx(styles.formInput)}
            fullWidth
            placeholder="+7 999 123 45 67"
            type="tel"
          />
          <FieldError />
        </TextField>
        <TextField
          className={clsx(styles.formField)}
          isRequired
          name="email"
          onChange={form.setEmail}
          type="email"
          validate={form.validateEmailField}
          value={form.email}
        >
          <Label className={clsx(styles.formLabel)}>Email</Label>
          <Input className={clsx(styles.formInput)} fullWidth placeholder="client@company.ru" />
          <FieldError />
        </TextField>
      </div>
      <div className={clsx(styles.formActions)}>
        <Button className={clsx(styles.formSubmit)} isDisabled={pending} type="submit">
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
  const form = usePasswordForm({ onSubmit });
  const passwordChangeVersion = useUnit($passwordChangeVersion);

  useEffect(() => {
    if (passwordChangeVersion === 0) {
      return;
    }

    form.reset();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- только после успешного ответа API
  }, [passwordChangeVersion]);

  return (
    <Form onSubmit={form.handleSubmit}>
      <div className={clsx(styles.formGrid)}>
        <TextField
          className={clsx(styles.formField)}
          isRequired
          name="oldPassword"
          onChange={form.setOldPassword}
          type="password"
          validate={form.validateOldPassword}
          value={form.oldPassword}
        >
          <Label className={clsx(styles.formLabel)}>Текущий пароль</Label>
          <Input className={clsx(styles.formInput)} fullWidth placeholder="••••••••" />
          <FieldError />
        </TextField>
        <TextField
          className={clsx(styles.formField)}
          isRequired
          minLength={6}
          name="newPassword"
          onChange={form.setNewPassword}
          type="password"
          validate={form.validateNewPassword}
          value={form.newPassword}
        >
          <Label className={clsx(styles.formLabel)}>Новый пароль</Label>
          <Input className={clsx(styles.formInput)} fullWidth placeholder="••••••••" />
          <FieldError />
        </TextField>
        <TextField
          className={clsx(styles.formField)}
          isRequired
          name="confirmPassword"
          onChange={form.setConfirmPassword}
          type="password"
          validate={form.validateConfirmPassword}
          value={form.confirmPassword}
        >
          <Label className={clsx(styles.formLabel)}>Повтор нового пароля</Label>
          <Input className={clsx(styles.formInput)} fullWidth placeholder="••••••••" />
          <FieldError />
        </TextField>
      </div>
      <div className={clsx(styles.formActions)}>
        <Button className={clsx(styles.formSubmitSecondary)} isDisabled={pending} type="submit">
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
    isPasswordPending,
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
    changePasswordFx.pending,
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
            <PasswordForm onSubmit={changePassword} pending={isPasswordPending} />
          </section>
        </Surface>
      ) : null}
    </div>
  );
};

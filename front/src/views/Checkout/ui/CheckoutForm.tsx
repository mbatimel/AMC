'use client';

import {
  Alert,
  Button,
  FieldError,
  Form,
  Input,
  Label,
  Radio,
  RadioGroup,
  Surface,
  TextArea,
  TextField,
  Typography,
} from '@heroui/react';
import clsx from 'clsx';
import { useState } from 'react';

import {
  formatPhoneInput,
  normalizePhone,
  validateEmail,
  validatePhone,
} from '@/core/shared/lib/validateContact';

import type { CheckoutFormValues, DeliveryType } from '../lib/types';

import styles from '../Checkout.module.css';
import { DELIVERY_OPTIONS, PICKUP_WAREHOUSE_ADDRESS } from '../lib/constants';

type CheckoutFormProps = {
  error: null | string;
  isPending: boolean;
  onSubmit: (values: CheckoutFormValues) => void;
};

export const CheckoutForm = ({ error, isPending, onSubmit }: CheckoutFormProps): JSX.Element => {
  const [deliveryType, setDeliveryType] = useState<DeliveryType>('pickup');
  const [deliveryAddress, setDeliveryAddress] = useState(PICKUP_WAREHOUSE_ADDRESS);
  const [contactName, setContactName] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [comment, setComment] = useState('');
  const [phoneError, setPhoneError] = useState<null | string>(null);
  const [emailError, setEmailError] = useState<null | string>(null);

  const isPickup = deliveryType === 'pickup';

  return (
    <Surface className={clsx(styles.form)}>
      <Form
        onSubmit={(event) => {
          event.preventDefault();

          const nextPhoneError = validatePhone(phone);
          const nextEmailError = validateEmail(email, { required: false });

          setPhoneError(nextPhoneError);
          setEmailError(nextEmailError);

          if (nextPhoneError || nextEmailError) {
            return;
          }

          onSubmit({
            comment,
            contactName,
            deliveryAddress: isPickup ? PICKUP_WAREHOUSE_ADDRESS : deliveryAddress,
            deliveryType,
            email: email.trim(),
            phone: normalizePhone(phone),
          });
        }}
      >
        <section className={clsx(styles.section)}>
          <Typography.Heading className={clsx(styles.sectionTitle)} level={2}>
            Способ получения
          </Typography.Heading>
          <RadioGroup
            aria-label="Способ получения"
            className={clsx(styles.deliveryOptions)}
            onChange={(value) => {
              const next = value as DeliveryType;

              setDeliveryType(next);

              if (next === 'pickup') {
                setDeliveryAddress(PICKUP_WAREHOUSE_ADDRESS);
              } else if (deliveryAddress === PICKUP_WAREHOUSE_ADDRESS) {
                setDeliveryAddress('');
              }
            }}
            value={deliveryType}
          >
            {DELIVERY_OPTIONS.map((option) => (
              <Radio
                className={clsx(styles.deliveryOption)}
                key={option.value}
                value={option.value}
              >
                <Radio.Control>
                  <Radio.Indicator />
                </Radio.Control>
                <Radio.Content>
                  <span className={clsx(styles.deliveryLabel)}>{option.label}</span>
                  <span className={clsx(styles.deliveryDescription)}>{option.description}</span>
                </Radio.Content>
              </Radio>
            ))}
          </RadioGroup>
        </section>

        <section className={clsx(styles.section)}>
          <Typography.Heading className={clsx(styles.sectionTitle)} level={2}>
            Адрес
          </Typography.Heading>
          <TextField
            isDisabled={isPickup}
            isRequired
            name="deliveryAddress"
            onChange={setDeliveryAddress}
            value={deliveryAddress}
          >
            <Label>{isPickup ? 'Адрес склада' : 'Адрес доставки'}</Label>
            <Input placeholder="Город, улица, дом" />
          </TextField>
        </section>

        <section className={clsx(styles.section)}>
          <Typography.Heading className={clsx(styles.sectionTitle)} level={2}>
            Контактные данные
          </Typography.Heading>
          <div className={clsx(styles.formGrid)}>
            <TextField isRequired name="contactName" onChange={setContactName} value={contactName}>
              <Label>Контактное лицо</Label>
              <Input placeholder="ФИО" />
            </TextField>
            <TextField
              isInvalid={Boolean(phoneError)}
              isRequired
              name="phone"
              onChange={(value) => {
                setPhone((previous) => formatPhoneInput(value, previous));
                setPhoneError(null);
              }}
              value={phone}
            >
              <Label>Телефон</Label>
              <Input placeholder="+7 999 123 45 67" type="tel" />
              {phoneError ? <FieldError>{phoneError}</FieldError> : null}
            </TextField>
            <TextField
              isInvalid={Boolean(emailError)}
              name="email"
              onChange={(value) => {
                setEmail(value);
                setEmailError(null);
              }}
              value={email}
            >
              <Label>Email</Label>
              <Input placeholder="name@company.ru" type="email" />
              {emailError ? <FieldError>{emailError}</FieldError> : null}
            </TextField>
          </div>
        </section>

        <section className={clsx(styles.section)}>
          <Typography.Heading className={clsx(styles.sectionTitle)} level={2}>
            Комментарий
          </Typography.Heading>
          <TextField name="comment" onChange={setComment} value={comment}>
            <Label>Пожелания к заказу</Label>
            <TextArea placeholder="Например, удобное время отгрузки" />
          </TextField>
        </section>

        {error ? (
          <Alert className={clsx(styles.alert)} status="danger">
            <Alert.Indicator />
            <Alert.Content>
              <Alert.Description>{error}</Alert.Description>
            </Alert.Content>
          </Alert>
        ) : null}
        {error ? <FieldError className={clsx(styles.srOnly)}>{error}</FieldError> : null}

        <div className={clsx(styles.formActions)}>
          <Button isDisabled={isPending} type="submit" variant="primary">
            {isPending ? 'Оформляем…' : 'Подтвердить заказ'}
          </Button>
        </div>
      </Form>
    </Surface>
  );
};

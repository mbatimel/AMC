'use client';

import {
  Alert,
  FieldError,
  Form,
  Input,
  Label,
  Radio,
  RadioGroup,
  TextArea,
  TextField,
} from '@heroui/react';
import clsx from 'clsx';
import { useState } from 'react';

import {
  formatPhoneInput,
  normalizePhone,
  validateEmail,
  validatePhone,
  validateRequired,
} from '@/core/shared/lib/validateContact';

import type { CheckoutFormValues, DeliveryType } from '../lib/types';

import styles from '../Checkout.module.css';
import { DELIVERY_OPTIONS, PICKUP_WAREHOUSE_ADDRESS } from '../lib/constants';

export const CHECKOUT_FORM_ID = 'checkout-form';

type CheckoutFormProps = {
  cityName: string;
  error: null | string;
  onSubmit: (values: CheckoutFormValues) => void;
};

export const CheckoutForm = ({ cityName, error, onSubmit }: CheckoutFormProps): JSX.Element => {
  const [deliveryType, setDeliveryType] = useState<DeliveryType>('pickup');
  const [deliveryAddress, setDeliveryAddress] = useState(PICKUP_WAREHOUSE_ADDRESS);
  const [contactName, setContactName] = useState('');
  const [phone, setPhoneValue] = useState('');
  const [email, setEmail] = useState('');
  const [comment, setComment] = useState('');

  const setPhone = (value: string): void => {
    setPhoneValue((previous) => formatPhoneInput(value, previous));
  };

  return (
    <Form
      className={clsx(styles.form)}
      id={CHECKOUT_FORM_ID}
      onSubmit={(event) => {
        event.preventDefault();

        onSubmit({
          comment,
          contactName: contactName.trim(),
          deliveryAddress: deliveryAddress.trim(),
          deliveryType,
          email: email.trim(),
          phone: normalizePhone(phone),
        });
      }}
    >
      <section className={clsx(styles.section)}>
        <h2 className={clsx(styles.sectionTitle)}>Способ доставки</h2>
        <RadioGroup
          aria-label="Способ доставки"
          className={clsx(styles.deliveryOptions)}
          onChange={(value) => {
            const next = value as DeliveryType;

            setDeliveryType(next);

            if (next === 'pickup' && deliveryAddress.trim().length === 0) {
              setDeliveryAddress(PICKUP_WAREHOUSE_ADDRESS);
            }
          }}
          value={deliveryType}
        >
          {DELIVERY_OPTIONS.map((option) => (
            <Radio className={clsx(styles.deliveryOption)} key={option.value} value={option.value}>
              <Radio.Control className={clsx(styles.deliveryControl)}>
                <Radio.Indicator />
              </Radio.Control>
              <span className={clsx(styles.deliveryLabel)}>{option.label(cityName)}</span>
            </Radio>
          ))}
        </RadioGroup>
      </section>

      <section className={clsx(styles.section)}>
        <h2 className={clsx(styles.sectionTitle)}>Адрес доставки</h2>
        <TextField
          className={clsx(styles.field)}
          isRequired
          name="deliveryAddress"
          onChange={setDeliveryAddress}
          validate={validateRequired}
          value={deliveryAddress}
        >
          <Label className={clsx(styles.fieldLabel)}>Адрес доставки</Label>
          <Input className={clsx(styles.fieldInput)} fullWidth placeholder="Город, улица, дом" />
          <FieldError />
        </TextField>
      </section>

      <section className={clsx(styles.section)}>
        <h2 className={clsx(styles.sectionTitle)}>Контактные данные</h2>
        <div className={clsx(styles.formGrid)}>
          <TextField
            className={clsx(styles.field)}
            isRequired
            name="contactName"
            onChange={setContactName}
            validate={validateRequired}
            value={contactName}
          >
            <Label className={clsx(styles.fieldLabel)}>Контактное лицо</Label>
            <Input className={clsx(styles.fieldInput)} fullWidth placeholder="ФИО" />
            <FieldError />
          </TextField>
          <TextField
            className={clsx(styles.field)}
            isRequired
            name="phone"
            onChange={setPhone}
            validate={validatePhone}
            value={phone}
          >
            <Label className={clsx(styles.fieldLabel)}>Телефон</Label>
            <Input
              className={clsx(styles.fieldInput)}
              fullWidth
              placeholder="+7 999 123 45 67"
              type="tel"
            />
            <FieldError />
          </TextField>
          <TextField
            className={clsx(styles.field, styles.emailField)}
            isRequired
            name="email"
            onChange={setEmail}
            type="email"
            validate={validateEmail}
            value={email}
          >
            <Label className={clsx(styles.fieldLabel)}>Email</Label>
            <Input
              className={clsx(styles.fieldInput)}
              fullWidth
              placeholder="name@company.ru"
              type="email"
            />
            <FieldError />
          </TextField>
        </div>
      </section>

      <section className={clsx(styles.section)}>
        <h2 className={clsx(styles.sectionTitle)}>Комментарий</h2>
        <TextField
          aria-label="Комментарий"
          className={clsx(styles.field)}
          name="comment"
          onChange={setComment}
          value={comment}
        >
          <TextArea
            className={clsx(styles.commentArea)}
            fullWidth
            placeholder="Комментарий к заказу (например: позвонить за 1 час до отгрузки)"
          />
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

      <div className={clsx(styles.infoBox)} role="note">
        После оформления заказ попадёт в личный кабинет. Счёт и документы появятся там же; передача
        в 1С выполняется на стороне сервиса.
      </div>
    </Form>
  );
};

const ORDER_STATUS_LABELS: Record<string, string> = {
  cancelled: 'Отменён',
  completed: 'Завершён',
  confirmed: 'Подтверждён',
  delivered: 'Доставлен',
  draft: 'Черновик',
  new: 'Новый',
  processing: 'На рассмотрении',
  shipped: 'Отгружен',
};

const PAYMENT_STATUS_LABELS: Record<string, string> = {
  not_paid: 'Не оплачен',
  paid: 'Оплачен',
  partially_paid: 'Частично оплачен',
  pending: 'Ожидает оплаты',
  unpaid: 'Не оплачен',
};

const DELIVERY_TYPE_LABELS: Record<string, string> = {
  courier: 'Курьер',
  pickup: 'Самовывоз',
  transport: 'Транспортная компания',
};

export const formatOrderStatus = (status: string): string =>
  ORDER_STATUS_LABELS[status] ?? (status || '—');

export const formatPaymentStatus = (status: string): string =>
  PAYMENT_STATUS_LABELS[status] ?? (status || '—');

export const formatDeliveryType = (type: string): string =>
  DELIVERY_TYPE_LABELS[type] ?? (type || '—');

export const formatOrderDate = (value: string): string => {
  if (!value) {
    return '—';
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(date);
};

export const statusChipColor = (
  status: string,
): 'accent' | 'danger' | 'default' | 'success' | 'warning' => {
  if (status === 'cancelled') {
    return 'danger';
  }

  if (status === 'completed' || status === 'delivered') {
    return 'default';
  }

  if (status === 'processing') {
    return 'warning';
  }

  if (status === 'shipped') {
    return 'success';
  }

  if (status === 'new' || status === 'confirmed') {
    return 'accent';
  }

  return 'default';
};

export const paymentChipColor = (
  status: string,
): 'accent' | 'danger' | 'default' | 'success' | 'warning' => {
  if (status === 'paid') {
    return 'success';
  }

  if (status === 'unpaid' || status === 'not_paid') {
    return 'danger';
  }

  if (status === 'partially_paid' || status === 'pending') {
    return 'warning';
  }

  return 'default';
};

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

export const formatOrderDateTime = (value: string): string => {
  if (!value) {
    return '—';
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(date);
};

export const statusTone = (
  status: string,
): 'cancelled' | 'completed' | 'default' | 'new' | 'processing' | 'shipped' => {
  if (status === 'cancelled') {
    return 'cancelled';
  }

  if (status === 'completed' || status === 'delivered') {
    return 'completed';
  }

  if (status === 'processing') {
    return 'processing';
  }

  if (status === 'shipped') {
    return 'shipped';
  }

  if (status === 'new' || status === 'confirmed') {
    return 'new';
  }

  return 'default';
};

export const paymentTone = (status: string): 'default' | 'paid' | 'pending' | 'unpaid' => {
  if (status === 'paid') {
    return 'paid';
  }

  if (status === 'unpaid' || status === 'not_paid') {
    return 'unpaid';
  }

  if (status === 'partially_paid' || status === 'pending') {
    return 'pending';
  }

  return 'default';
};

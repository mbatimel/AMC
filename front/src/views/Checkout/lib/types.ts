export type CheckoutFormValues = {
  comment: string;
  contactName: string;
  deliveryAddress: string;
  deliveryType: DeliveryType;
  email: string;
  phone: string;
};

export type DeliveryType = 'courier' | 'pickup' | 'transport';

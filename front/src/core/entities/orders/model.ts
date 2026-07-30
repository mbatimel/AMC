import { createEvent } from 'effector';

/** Сброс кэша списка заказов (после CreateOrder и т.п.). */
export const ordersListInvalidated = createEvent();

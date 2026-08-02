/**
 * Типы данных портальных модулей, у которых пока нет собственного Go-сервиса
 * (контент страниц, баннеры, юр. документы, заявки, отзывы, обращения, чат).
 *
 * Контракт намеренно повторяет форму ответов backend-сервисов
 * (`{ error, errorText, data }`), чтобы при появлении реального сервиса
 * достаточно было поменять базовый URL в `core/shared/api/*`.
 */

export type AuditLogEntry = {
  action: string;
  actor_label: string;
  created_at: string;
  id: string;
};

export type BannerItem = {
  dateFrom: string;
  dateTo: string;
  id: string;
  image: string;
  is_active: boolean;
  link: string;
  sort_order: number;
  subtitle: string;
  title: string;
};

export type BannersSettings = {
  delay_sec: number;
  items: BannerItem[];
};

export type ContactsPageContent = {
  address: string;
  email: string;
  managers: { email: string; name: string; phone: string; role: string }[];
  phone: string;
  requisites: string;
  title: string;
  work_hours: string;
};

export type ContentPageKey = 'about' | 'certificates' | 'contacts' | 'home' | 'promo' | 'terms';

export type ContentPages = {
  about: TextPageContent;
  certificates: ListPageContent;
  contacts: ContactsPageContent;
  home: HomePageContent;
  promo: ListPageContent;
  terms: TextPageContent;
};

export type HomePageContent = {
  features: string[];
  hero_button: string;
  hero_subtitle: string;
  hero_title: string;
  stats: { label: string; value: string }[];
  why_items: string[];
  why_title: string;
};

export type LegalDoc = {
  body: string;
  current_version: string;
  id: string;
  name: string;
  updated_at: string;
  versions: LegalDocVersion[];
};

export type LegalDocVersion = {
  author: string;
  date: string;
  summary: string;
  version: string;
};

export type ListPageContent = {
  items: string[];
  text: string;
  title: string;
};

export type OrderFeedback = {
  client_name: string;
  created_at: string;
  id: string;
  order_id: string;
  order_status: string;
  rating: number;
  text: string;
  user_id: string;
};

export type PortalState = {
  audit_log: AuditLogEntry[];
  banners: BannersSettings;
  content: ContentPages;
  feedback: OrderFeedback[];
  legal_docs: LegalDoc[];
  portal_users: PortalUser[];
  promotions: Promotion[];
  signup_requests: SignupRequest[];
  support_requests: SupportRequest[];
};

export type PortalUser = {
  company: string;
  contact: string;
  created_at: string;
  email: string;
  id: string;
  inn: string;
  is_active: boolean;
  phone: string;
  role: 'admin' | 'client';
};

export type Promotion = {
  condition: string;
  createdAt: string;
  desc: string;
  discMode: PromotionDiscMode;
  discValue: number;
  endAt: string;
  endedManually: boolean;
  id: string;
  minQty: number;
  name: string;
  sel: PromotionSelection;
  startAt: string;
  type: PromotionType;
};

export type PromotionDiscMode = 'percent' | 'price';

export type PromotionSelection = {
  all: boolean;
  nodes: string[];
  products: string[];
};

export type PromotionType = 'date' | 'qty';

export type SignupRequest = {
  company: string;
  contact: string;
  created_at: string;
  email: string;
  id: string;
  inn: string;
  phone: string;
  reject_reason: string;
  status: SignupRequestStatus;
  type: 'individual' | 'organization';
};

export type SignupRequestStatus = 'approved' | 'pending' | 'rejected';

export type SupportRequest = {
  answer: string;
  client_name: string;
  contact: string;
  created_at: string;
  id: string;
  order_id: string;
  severity: number;
  source: SupportRequestSource;
  status: SupportRequestStatus;
  subject: string;
  text: string;
  user_id: string;
};

export type SupportRequestSource = 'assistant' | 'form';

export type SupportRequestStatus = 'closed' | 'in_progress' | 'new';

export type TextPageContent = {
  text: string;
  title: string;
};

/**
 * Типы данных портальных модулей, у которых пока нет собственного Go-сервиса
 * (контент страниц, баннеры, юр. документы, заявки, отзывы, обращения, чат).
 *
 * Контракт намеренно повторяет форму ответов backend-сервисов
 * (`{ error, errorText, data }`), чтобы при появлении реального сервиса
 * достаточно было поменять базовый URL в `core/shared/api/*`.
 */

export type AboutPageContent = TextPageContent & {
  cta_badge?: string;
  cta_button?: string;
  cta_hint?: string;
  cta_text?: string;
  cta_title?: string;
  directions_badge?: string;
  directions_subtitle?: string;
  directions_title?: string;
  hero_badge?: string;
  hero_subtitle?: string;
  offices_badge?: string;
  offices_subtitle?: string;
  offices_title?: string;
  profile_badge?: string;
  profile_title?: string;
};

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

export type ContactOffice = {
  address: string;
  city: string;
  email: string;
  is_main?: boolean;
  phone: string;
};

export type ContactRequisiteItem = {
  label: string;
  value: string;
};

export type ContactsPageContent = {
  address: string;
  email: string;
  managers: { email: string; name: string; phone: string; role: string }[];
  offices: ContactOffice[];
  phone: string;
  requisite_items: ContactRequisiteItem[];
  /** Краткое резюме реквизитов (для обратной совместимости / админки). */
  requisites: string;
  subtitle: string;
  title: string;
  /** Подпись под картой склада. */
  warehouse_map_caption?: string;
  /** URL виджета Яндекс.Карт (iframe src). */
  warehouse_map_url?: string;
  work_hours: string;
};

export type ContentPageKey = 'about' | 'certificates' | 'contacts' | 'home' | 'promo' | 'terms';

export type ContentPages = {
  about: AboutPageContent;
  certificates: ListPageContent;
  contacts: ContactsPageContent;
  home: HomePageContent;
  promo: ListPageContent;
  terms: TextPageContent;
};

export type EmailLogEntry = {
  created_at: string;
  id: string;
  subject: string;
  text: string;
  to: string;
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
  email_log: EmailLogEntry[];
  feedback: OrderFeedback[];
  legal_docs: LegalDoc[];
  portal_users: PortalUser[];
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

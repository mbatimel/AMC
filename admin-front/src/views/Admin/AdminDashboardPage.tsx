'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import { useContent } from '@/core/entities/content';
import { AppPath, getContentPath } from '@/core/shared/router/paths';

import styles from './Admin.module.css';
import { formatAdminDateTime } from './lib/nav';
import { $auditLog, adminAuditLogOpened } from './model/audit';
import { $adminProducts, $adminProductsTotal, adminCatalogOpened } from './model/catalog';
import { $adminSupport, adminSupportOpened } from './model/feedback';
import { $signupRequests, adminUsersOpened } from './model/users';
import { AdminPageHeader } from './ui/AdminPageHeader';

export const AdminDashboardPage = (): JSX.Element => {
  const { banners, legalDocs } = useContent();
  const [
    products,
    productsTotal,
    signupRequests,
    supportRequests,
    auditLog,
    openCatalog,
    openUsers,
    openSupport,
    openAudit,
  ] = useUnit([
    $adminProducts,
    $adminProductsTotal,
    $signupRequests,
    $adminSupport,
    $auditLog,
    adminCatalogOpened,
    adminUsersOpened,
    adminSupportOpened,
    adminAuditLogOpened,
  ]);

  useEffect(() => {
    openCatalog();
    openUsers();
    openSupport();
    openAudit();
  }, [openAudit, openCatalog, openSupport, openUsers]);

  const withoutPhoto = products.filter((product) => (product.images ?? []).length === 0).length;
  const unpublished = products.filter((product) => product.is_published === false).length;
  const activeBanners = (banners?.items ?? []).filter((item) => item.is_active).length;
  const pendingSignups = signupRequests.filter((item) => item.status === 'pending').length;
  const openSupportCount = supportRequests.filter((item) => item.status !== 'closed').length;

  return (
    <>
      <AdminPageHeader
        subtitle="Администратор портала · что требует внимания"
        title="Сводка"
      />

      <div className={clsx(styles.kpiGrid)}>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Товары без фото</p>
          <p className={clsx(styles.kpiValue, withoutPhoto > 0 && styles.kpiValueAlert)}>
            {withoutPhoto}
          </p>
          <p className={clsx(styles.kpiHint)}>из {productsTotal} позиций каталога</p>
        </article>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Снято с публикации</p>
          <p className={clsx(styles.kpiValue)}>{unpublished}</p>
          <p className={clsx(styles.kpiHint)}>не видны клиентам</p>
        </article>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Активные баннеры</p>
          <p className={clsx(styles.kpiValue)}>{activeBanners}</p>
          <p className={clsx(styles.kpiHint)}>на главной странице</p>
        </article>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Заявки на регистрацию</p>
          <p className={clsx(styles.kpiValue, pendingSignups > 0 && styles.kpiValueAlert)}>
            {pendingSignups}
          </p>
          <p className={clsx(styles.kpiHint)}>ожидают решения</p>
        </article>
        <article className={clsx(styles.kpi)}>
          <p className={clsx(styles.kpiLabel)}>Обращения поддержки</p>
          <p className={clsx(styles.kpiValue, openSupportCount > 0 && styles.kpiValueAlert)}>
            {openSupportCount}
          </p>
          <p className={clsx(styles.kpiHint)}>не закрыты</p>
        </article>
      </div>

      <div className={clsx(styles.grid2)}>
        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Юридические документы</h2>
          {legalDocs.length === 0 ? (
            <p className={clsx(styles.hint)}>Документы не загружены</p>
          ) : (
            <table className={clsx(styles.table)}>
              <tbody>
                {legalDocs.map((doc) => (
                  <tr key={doc.id}>
                    <td>{doc.name}</td>
                    <td>
                      <span className={clsx(styles.badge, styles.badgeSuccess)}>
                        v{doc.current_version}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <p className={clsx(styles.hint)}>
            <Link href={AppPath.Legal}>Открыть редактор документов →</Link>
          </p>
        </section>

        <section className={clsx(styles.card)}>
          <h2 className={clsx(styles.cardTitle)}>Последние действия</h2>
          {auditLog.length === 0 ? (
            <p className={clsx(styles.hint)}>Записей пока нет</p>
          ) : (
            <table className={clsx(styles.table)}>
              <tbody>
                {auditLog.slice(0, 6).map((entry) => (
                  <tr key={entry.id}>
                    <td>{formatAdminDateTime(entry.createdAt)}</td>
                    <td>{entry.action}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <p className={clsx(styles.hint)}>
            <Link href={AppPath.AuditLog}>Открыть журнал →</Link>
          </p>
        </section>
      </div>

      <section className={clsx(styles.card)}>
        <h2 className={clsx(styles.cardTitle)}>Быстрые переходы</h2>
        <div className={clsx(styles.actionsRow)}>
          <Link className={clsx(styles.smallButton)} href={getContentPath('home')}>
            Главная страница
          </Link>
          <Link className={clsx(styles.smallButton)} href={AppPath.Banners}>
            Баннеры
          </Link>
          <Link className={clsx(styles.smallButton)} href={AppPath.Products}>
            Товары
          </Link>
          <Link className={clsx(styles.smallButton)} href={AppPath.SignupRequests}>
            Заявки
          </Link>
          <Link className={clsx(styles.smallButton)} href={AppPath.Support}>
            Обращения
          </Link>
        </div>
      </section>
    </>
  );
};

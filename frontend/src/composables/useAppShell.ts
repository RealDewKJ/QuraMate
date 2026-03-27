import { useI18n } from 'vue-i18n';

import { useAppBootstrap } from './useAppBootstrap';
import { useAppConnectionSession } from './useAppConnectionSession';

export function useAppShell() {
  const { locale, t } = useI18n({ useScope: 'global' });
  const { initializeAppBootstrap, trustSqlServerCertificateByDefault } = useAppBootstrap({ locale });

  const connectionSession = useAppConnectionSession({
    t,
    getTrustSqlServerCertificateByDefault: () => trustSqlServerCertificateByDefault.value,
  });

  const initializeAppShell = async () => {
    await initializeAppBootstrap();
    await connectionSession.initializeConnectionSession();
  };

  return {
    initializeAppShell,
    t,
    ...connectionSession,
  };
}

import type { ConnectionConfig } from './useConnectionForm';

export const APP_CONNECTION_SESSION_KEY = 'app_connection_session_v1';
export const APP_CONNECTION_SESSION_VERSION = 1;
export const APP_CONNECTION_RECOVERY_PREF_KEY = 'app_connection_recovery_pref_v1';
export const APP_CONNECTION_RECOVERY_PREF_VERSION = 1;

export interface PersistedConnectionEntry {
  name: string;
  sessionKey: string;
  config: ConnectionConfig;
}

export interface PersistedConnectionSession {
  version: number;
  activeSessionKey: string | null;
  connections: PersistedConnectionEntry[];
}

export interface RecoveryCandidate extends PersistedConnectionEntry {
  selected: boolean;
}

export interface RecoveryPreference {
  version: number;
  autoRecover: boolean;
  selectedSessionKeys: string[];
}

export const hashString = (value: string): string => {
  let hash = 5381;
  for (let index = 0; index < value.length; index += 1) {
    hash = ((hash << 5) + hash) ^ value.charCodeAt(index);
  }
  return (hash >>> 0).toString(16);
};

export const buildConnectionSessionKey = (config: Partial<ConnectionConfig>, displayName?: string): string => {
  const type = String(config.type || '').toLowerCase();
  const normalizedName = String(displayName || config.name || '').trim().toLowerCase();

  if (type === 'sqlite' || type === 'duckdb' || type === 'libsql') {
    const id = `file|${type}|${String(config.database || '').trim().toLowerCase()}|${normalizedName}`;
    return `conn_${hashString(id)}`;
  }

  const sshSection = config.sshEnabled
    ? `|ssh|${String(config.sshHost || '').trim().toLowerCase()}|${String(config.sshPort || 22)}|${String(config.sshUser || '').trim().toLowerCase()}|${String(config.sshKeyFile || '').trim().toLowerCase()}`
    : '|ssh|off';

  const id = [
    'net',
    type,
    String(config.host || '').trim().toLowerCase(),
    String(config.port || ''),
    String(config.user || '').trim().toLowerCase(),
    String(config.database || '').trim().toLowerCase(),
    normalizedName,
    sshSection,
  ].join('|');

  return `conn_${hashString(id)}`;
};

export const getDefaultConnectionConfig = (getTrustSqlServerCertificateByDefault: () => boolean): ConnectionConfig => ({
  id: '',
  name: '',
  type: 'postgres',
  host: 'localhost',
  port: 5432,
  user: 'postgres',
  password: '',
  database: 'postgres',
  readOnly: false,
  trustServerCertificate: getTrustSqlServerCertificateByDefault(),
  sshEnabled: false,
  sshHost: '',
  sshPort: 22,
  sshUser: '',
  sshPassword: '',
  sshKeyFile: '',
  encoding: '',
});

export const getNormalizedTrustServerCertificate = (
  config: Partial<ConnectionConfig>,
  getTrustSqlServerCertificateByDefault: () => boolean,
): boolean => {
  if ((config.type || '').toLowerCase() !== 'mssql') {
    return false;
  }

  if (typeof config.trustServerCertificate === 'boolean') {
    return config.trustServerCertificate;
  }

  return getTrustSqlServerCertificateByDefault();
};

export const normalizeConnectionConfig = (
  config: Partial<ConnectionConfig>,
  getTrustSqlServerCertificateByDefault: () => boolean,
): ConnectionConfig => ({
  ...getDefaultConnectionConfig(getTrustSqlServerCertificateByDefault),
  ...config,
  type: config.type || 'postgres',
  host: config.host || '',
  port: typeof config.port === 'number' ? config.port : Number(config.port || 0),
  user: config.user || '',
  password: config.password || '',
  database: config.database || '',
  readOnly: !!config.readOnly,
  trustServerCertificate: getNormalizedTrustServerCertificate(config, getTrustSqlServerCertificateByDefault),
  sshEnabled: !!config.sshEnabled,
  sshHost: config.sshHost || '',
  sshPort: typeof config.sshPort === 'number' ? config.sshPort : Number(config.sshPort || 22),
  sshUser: config.sshUser || '',
  sshPassword: config.sshPassword || '',
  sshKeyFile: config.sshKeyFile || '',
  encoding: config.encoding || '',
  name: config.name || '',
  id: config.id || '',
});

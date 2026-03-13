import { ref, reactive, computed, watch } from "vue";
import { useLocalStorage, useToggle, watchImmediate } from "@vueuse/core";
import {
  ConnectDB,
  GetSSHHostKeyInfo,
  ReadTextFile,
  SelectImportFile,
  TestConnection,
  SetReadOnly,
  SelectSqliteFile,
  SaveCredential,
  DeleteCredential,
  SelectExportFile,
  TrustSSHHost,
  WriteTextFile,
} from "../../wailsjs/go/app/App";
import { app } from "../../wailsjs/go/models";

// ---------- types ----------

export interface ConnectionConfig {
  id: string;
  name: string;
  type: string;
  host: string;
  port: number;
  user: string;
  password: string;
  database: string;
  readOnly: boolean;
  trustServerCertificate: boolean;
  sshEnabled: boolean;
  sshHost: string;
  sshPort: number;
  sshUser: string;
  sshPassword: string;
  sshKeyFile: string;
  encoding: string;
}

export interface ActiveConnection {
  id: string;
  name: string;
  config: ConnectionConfig;
}

export interface SSHTrustAuditEntry {
  host: string;
  port: number;
  pattern: string;
  keyType: string;
  fingerprint: string;
  trustedAt: string;
  rotatedFromFingerprint?: string;
  rotationReason?: string;
}

type ConnectionFieldKey =
  | "host"
  | "port"
  | "user"
  | "database"
  | "sshHost"
  | "sshPort"
  | "sshUser"
  | "sshAuth";

export type ConnectionFieldErrors = Partial<Record<ConnectionFieldKey, string>>;

// ---------- helpers ----------

const DEFAULT_CONFIG: ConnectionConfig = {
  id: "",
  name: "",
  type: "postgres",
  host: "localhost",
  port: 5432,
  user: "postgres",
  password: "",
  database: "postgres",
  readOnly: false,
  trustServerCertificate: true,
  sshEnabled: false,
  sshHost: "",
  sshPort: 22,
  sshUser: "",
  sshPassword: "",
  sshKeyFile: "",
  encoding: "",
};

const LOCAL_DATABASE_TYPES = new Set(["sqlite", "duckdb", "libsql"]);
const ERROR_CODE_PATTERN = /^Error \[([A-Z0-9_]+)\]:/;

const isLocalDatabaseType = (type: string): boolean =>
  LOCAL_DATABASE_TYPES.has((type || "").toLowerCase());

const getDefaultPortByType = (type: string): number | null => {
  switch ((type || "").toLowerCase()) {
    case "postgres":
    case "greenplum":
    case "redshift":
      return 5432;
    case "cockroachdb":
      return 26257;
    case "mysql":
    case "mariadb":
      return 3306;
    case "databend":
      return 3307;
    case "mssql":
      return 1433;
    default:
      return null;
  }
};

const extractErrorCode = (message: string): string => {
  const match = (message || "").match(ERROR_CODE_PATTERN);
  return match?.[1] || "";
};

const buildFieldErrors = (target: ConnectionConfig): ConnectionFieldErrors => {
  const errors: ConnectionFieldErrors = {};
  const isLocalType = isLocalDatabaseType(target.type);

  if (!isLocalType) {
    if (!target.host.trim()) errors.host = "Host is required.";
    if (!target.user.trim()) errors.user = "User is required.";
    if (!target.database.trim()) errors.database = "Database name is required.";
    if (!Number.isInteger(target.port) || target.port < 1 || target.port > 65535) {
      errors.port = "Port must be between 1 and 65535.";
    }
  } else if (!target.database.trim()) {
    errors.database = "Database file path is required.";
  }

  if (target.sshEnabled && !isLocalType) {
    if (!target.sshHost.trim()) errors.sshHost = "SSH host is required.";
    if (!target.sshUser.trim()) errors.sshUser = "SSH user is required.";
    if (!Number.isInteger(target.sshPort) || target.sshPort < 1 || target.sshPort > 65535) {
      errors.sshPort = "SSH port must be between 1 and 65535.";
    }
    if (!target.sshPassword && !target.sshKeyFile.trim()) {
      errors.sshAuth = "SSH password or SSH key file is required.";
    }
  }

  return errors;
};

export function getConnectionLabel(conn: Partial<ConnectionConfig>): string {
  if (conn.name) return conn.name;
  if (
    conn.type === "sqlite" ||
    conn.type === "duckdb" ||
    conn.type === "libsql"
  )
    return `${conn.type === "sqlite" ? "SQLite" : conn.type === "duckdb" ? "DuckDB" : "LibSQL"}: ${conn.database}`;
  if (conn.type === "greenplum")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (Greenplum)`;
  if (conn.type === "redshift")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (Redshift)`;
  if (conn.type === "cockroachdb")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (CockroachDB)`;
  if (conn.type === "databend")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (Databend)`;
  return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (${conn.type})`;
}

export function isConfigEqual(
  c1: Partial<ConnectionConfig> | null,
  c2: Partial<ConnectionConfig> | null,
): boolean {
  if (!c1 || !c2) return false;

  const type1 = (c1.type || "").toLowerCase();
  const type2 = (c2.type || "").toLowerCase();
  if (type1 !== type2) return false;

  const db1 = c1.database || "";
  const db2 = c2.database || "";
  const name1 = c1.name || "";
  const name2 = c2.name || "";

  if (type1 === "sqlite" || type1 === "duckdb" || type1 === "libsql") {
    return db1 === db2 && name1 === name2;
  }

  const host1 = (c1.host || "").toLowerCase();
  const host2 = (c2.host || "").toLowerCase();
  const port1 = parseInt(String(c1.port), 10);
  const port2 = parseInt(String(c2.port), 10);
  const user1 = c1.user || "";
  const user2 = c2.user || "";

  if (
    host1 !== host2 ||
    port1 !== port2 ||
    user1 !== user2 ||
    db1 !== db2 ||
    name1 !== name2
  ) {
    return false;
  }

  const sshEnabled1 = !!c1.sshEnabled;
  const sshEnabled2 = !!c2.sshEnabled;
  if (sshEnabled1 !== sshEnabled2) return false;
  if (!!c1.trustServerCertificate !== !!c2.trustServerCertificate) return false;

  if (sshEnabled1) {
    const sshHost1 = (c1.sshHost || "").toLowerCase();
    const sshHost2 = (c2.sshHost || "").toLowerCase();
    const sshPort1 = parseInt(String(c1.sshPort || 22), 10);
    const sshPort2 = parseInt(String(c2.sshPort || 22), 10);
    const sshUser1 = c1.sshUser || "";
    const sshUser2 = c2.sshUser || "";
    const sshKeyFile1 = c1.sshKeyFile || "";
    const sshKeyFile2 = c2.sshKeyFile || "";

    return (
      sshHost1 === sshHost2 &&
      sshPort1 === sshPort2 &&
      sshUser1 === sshUser2 &&
      sshKeyFile1 === sshKeyFile2
    );
  }

  return true;
}

// ---------- composable ----------

export function useConnectionForm(
  activeConnections: () => ActiveConnection[],
  emitConnected: (conn: ActiveConnection) => void,
  emitConnectionExists: (id: string) => void,
  emitConnectionUpdated: (update: { id: string; config: ConnectionConfig }) => void,
  getTrustServerCertificateDefault: () => boolean = () => true,
) {
  const buildDefaultConfig = (): ConnectionConfig => ({
    ...DEFAULT_CONFIG,
    trustServerCertificate: getTrustServerCertificateDefault(),
  });

  const config = reactive<ConnectionConfig>(buildDefaultConfig());

  const resetConfig = () => {
    Object.assign(config, buildDefaultConfig());
  };

  const error = ref("");
  const lastErrorCode = ref("");
  const testSuccess = ref("");
  const isTrustingHost = ref(false);
  const isLoadingHostKeyInfo = ref(false);
    const sshHostKeyInfo = ref<app.SSHHostKeyInfo | null>(null);
  const expectedSshFingerprint = ref("");
  const sshRotationReason = ref("");
  const sshRotationConfirmText = ref("");
  const connectionToken = ref(0);
  const [isLoading] = useToggle(false);
  const [isTesting] = useToggle(false);
  const [isQuickConnecting] = useToggle(false);
  const savedConnections = useLocalStorage<ConnectionConfig[]>("savedConnections", []);
  const sshTrustAudit = useLocalStorage<SSHTrustAuditEntry[]>("sshTrustAudit", []);
  const sshTrustAuditSearch = ref("");

  const clearError = () => {
    error.value = "";
    lastErrorCode.value = "";
  };
  const setError = (message: string) => {
    error.value = message;
    lastErrorCode.value = extractErrorCode(message);
  };
  const applySqlServerTrustDefault = (
    target: ConnectionConfig,
  ): ConnectionConfig => {
    if ((target.type || "").toLowerCase() === "mssql") {
      return {
        ...target,
        trustServerCertificate: getTrustServerCertificateDefault(),
      };
    }
    return {
      ...target,
      trustServerCertificate: false,
    };
  };

  const connectionLabel = computed(() => getConnectionLabel(config));
  const fieldErrors = computed<ConnectionFieldErrors>(() => buildFieldErrors(config));
  const canTrustCurrentSshHost = computed(() => {
    if (!config.sshEnabled || isLocalDatabaseType(config.type)) return false;
    if (!config.sshHost.trim()) return false;
    return (
      lastErrorCode.value === "SSH_HOST_UNTRUSTED" ||
      lastErrorCode.value === "SSH_HOST_KEY_MISMATCH"
    );
  });
  const filteredSshTrustAudit = computed(() => {
    const query = sshTrustAuditSearch.value.trim().toLowerCase();
    if (!query) return sshTrustAudit.value;

    return sshTrustAudit.value.filter((entry) => {
      const haystack = `${entry.pattern} ${entry.fingerprint} ${entry.keyType} ${entry.host}:${entry.port}`.toLowerCase();
      return haystack.includes(query);
    });
  });

  const normalizeFingerprint = (value: string): string => value.trim().toLowerCase();
  const removeSHA256Prefix = (value: string): string => value.replace(/^sha256:/i, "");
  const normalizeForCompare = (value: string): string => removeSHA256Prefix(normalizeFingerprint(value));
  const currentSshPattern = computed(() => {
    const host = config.sshHost.trim();
    if (!host) return "";
    const port = Number.isInteger(config.sshPort) && config.sshPort > 0 ? config.sshPort : 22;
    if (port === 22) return host;
    if (host.includes(":") && !host.startsWith("[")) return `[${host}]:${port}`;
    return `${host}:${port}`;
  });
  const pinnedSshTrustEntry = computed(() => {
    const pattern = currentSshPattern.value;
    if (!pattern) return null;
    return sshTrustAudit.value.find((entry) => entry.pattern === pattern) || null;
  });
  const pinnedSshFingerprint = computed(() => pinnedSshTrustEntry.value?.fingerprint || "");

  const isFingerprintMatch = computed(() => {
    if (!sshHostKeyInfo.value) return false;
    const expected = normalizeFingerprint(expectedSshFingerprint.value);
    if (!expected) return true;

    const actual = normalizeFingerprint(sshHostKeyInfo.value.fingerprint || "");
    if (expected === actual) return true;

    return removeSHA256Prefix(expected) === removeSHA256Prefix(actual);
  });

  const isFingerprintMismatch = computed(() => {
    if (!sshHostKeyInfo.value) return false;
    return !!expectedSshFingerprint.value.trim() && !isFingerprintMatch.value;
  });
  const isPinnedFingerprintMismatch = computed(() => {
    if (!sshHostKeyInfo.value || !pinnedSshFingerprint.value) return false;
    return normalizeForCompare(sshHostKeyInfo.value.fingerprint) !== normalizeForCompare(pinnedSshFingerprint.value);
  });

  watch(
    () => [config.sshEnabled, config.sshHost, config.sshPort],
    () => {
      sshHostKeyInfo.value = null;
      expectedSshFingerprint.value = "";
      sshRotationReason.value = "";
      sshRotationConfirmText.value = "";
    },
  );

  // Auto-adjust port when db type changes
  watchImmediate(
    () => config.type,
    (newType, oldType) => {
      const nextDefaultPort = getDefaultPortByType(newType);
      const previousDefaultPort = getDefaultPortByType(oldType || "");

      if (
        nextDefaultPort !== null &&
        (
          !Number.isInteger(config.port) ||
          config.port <= 0 ||
          oldType === undefined ||
          (previousDefaultPort !== null && config.port === previousDefaultPort)
        )
      ) {
        config.port = nextDefaultPort;
      }

      if (isLocalDatabaseType(newType)) {
        config.sshEnabled = false;
        config.sshHost = "";
        config.sshPort = 22;
        config.sshUser = "";
        config.sshPassword = "";
        config.sshKeyFile = "";
      }
      if (newType === "mssql") {
        config.trustServerCertificate = getTrustServerCertificateDefault();
      } else {
        config.trustServerCertificate = false;
      }
    },
  );

  const validateConnectionConfig = (target: ConnectionConfig): string => {
    const errors = buildFieldErrors(target);
    return (
      errors.host ||
      errors.port ||
      errors.user ||
      errors.database ||
      errors.sshHost ||
      errors.sshPort ||
      errors.sshUser ||
      errors.sshAuth ||
      ""
    );
  };

  const handleSelectSqliteFile = async () => {
    try {
      const filePath = await SelectSqliteFile();
      if (filePath) {
        config.database = filePath;
      }
    } catch (e) {
      console.error("Failed to select SQLite file", e);
    }
  };

  const cancelConnection = () => {
    connectionToken.value++;
    isLoading.value = false;
    isTesting.value = false;
    isQuickConnecting.value = false;
    setError("Connection cancelled by user.");
    testSuccess.value = "";
  };

  const saveConnection = async (newConfig: ConnectionConfig) => {
    const normalizedConfig = applySqlServerTrustDefault({ ...newConfig });

    if (!normalizedConfig.id) {
      const existing = savedConnections.value.find((c) =>
        isConfigEqual(c, normalizedConfig),
      );
      if (existing && existing.id) {
        normalizedConfig.id = existing.id;
      } else {
        normalizedConfig.id = crypto.randomUUID();
      }
    }

    config.id = normalizedConfig.id;

    if (normalizedConfig.password || normalizedConfig.sshPassword) {
      try {
        await SaveCredential(
          normalizedConfig.id,
          normalizedConfig.password || "",
          normalizedConfig.sshPassword || "",
        );
      } catch (e) {
        console.error("Failed to save credentials to keyring", e);
      }
    }

    const storageConfig: ConnectionConfig = {
      ...normalizedConfig,
      password: "",
      sshPassword: "",
    };

    const existsIndex = savedConnections.value.findIndex(
      (c) => c.id === storageConfig.id || isConfigEqual(c, storageConfig),
    );

    if (existsIndex === -1) {
      savedConnections.value.push(storageConfig);
    } else {
      savedConnections.value[existsIndex] = storageConfig;
    }
  };

  const performConnect = async () => {
    clearError();
    testSuccess.value = "";
    sshHostKeyInfo.value = null;
    const effectiveConfig = applySqlServerTrustDefault({ ...config });

    const validationError = validateConnectionConfig(effectiveConfig);
    if (validationError) {
      setError(validationError);
      return;
    }

    isLoading.value = true;
    const currentToken = ++connectionToken.value;

    // Check for existing connection
    const existing = activeConnections().find((c) =>
      isConfigEqual(c.config, effectiveConfig),
    );

    if (existing) {
      if (currentToken !== connectionToken.value) return;
      // Check if ReadOnly status has changed
      if (!!existing.config.readOnly !== !!config.readOnly) {
        try {
          await SetReadOnly(existing.id, !!config.readOnly);
          emitConnectionUpdated({
            id: existing.id,
            config: { ...existing.config, readOnly: !!config.readOnly },
          });
        } catch (e: any) {
          console.error("Failed to update ReadOnly status:", e);
        }
      }

      emitConnectionExists(existing.id);
      resetConfig();
      isLoading.value = false;
      return;
    }

    try {
      const result = await ConnectDB(effectiveConfig);
      if (currentToken !== connectionToken.value) return;

      if (result.id) {
        await saveConnection(effectiveConfig);
        if (currentToken !== connectionToken.value) return;

        emitConnected({
          id: result.id,
          name: effectiveConfig.name || getConnectionLabel(effectiveConfig),
          config: effectiveConfig,
        });
        resetConfig();
      } else {
        setError(result.error || "Unknown error");
      }
    } catch (e: any) {
      if (currentToken !== connectionToken.value) return;
      setError(e.toString());
    } finally {
      if (currentToken === connectionToken.value) {
        isLoading.value = false;
        isQuickConnecting.value = false;
      }
    }
  };

  const connect = () => {
    performConnect();
  };

  const testConnection = async () => {
    clearError();
    testSuccess.value = "";
    sshHostKeyInfo.value = null;
    const effectiveConfig = applySqlServerTrustDefault({ ...config });

    const validationError = validateConnectionConfig(effectiveConfig);
    if (validationError) {
      setError(validationError);
      return;
    }

    isTesting.value = true;
    const currentToken = ++connectionToken.value;

    try {
      const testConfig = { ...effectiveConfig };
      if (!testConfig.id) {
        testConfig.id = "";
      }
      const result = await TestConnection(testConfig);
      if (currentToken !== connectionToken.value) return;

      if (result === "Success") {
        testSuccess.value = "Connection successful!";
      } else {
        setError(result);
      }
    } catch (e: any) {
      if (currentToken !== connectionToken.value) return;
      setError(e.toString());
    } finally {
      if (currentToken === connectionToken.value) {
        isTesting.value = false;
      }
    }
  };

  const removeConnection = async (index: number) => {
    const conn = savedConnections.value[index];
    if (conn && conn.id) {
      try {
        await DeleteCredential(conn.id);
      } catch (e) {
        console.error("Failed to delete credentials from keyring", e);
      }
    }
    savedConnections.value.splice(index, 1);
  };

  const selectConnection = (conn: ConnectionConfig) => {
    Object.assign(
      config,
      applySqlServerTrustDefault({
        ...DEFAULT_CONFIG,
        ...conn,
        name: conn.name || "",
      }),
    );
    config.password = "";
    config.sshPassword = "";

    isQuickConnecting.value = true;
    connect();
  };

  const editConnection = (conn: ConnectionConfig) => {
    Object.assign(
      config,
      applySqlServerTrustDefault({
        ...DEFAULT_CONFIG,
        ...conn,
        name: conn.name || "",
      }),
    );
    config.password = "";
    config.sshPassword = "";
  };

  const migrateSavedConnections = async () => {
    let needsSave = false;
    const connections = [...savedConnections.value];

    for (const conn of connections) {
      let modified = false;
      if (!conn.id) {
        conn.id = crypto.randomUUID();
        modified = true;
      }
      const normalizedConn = applySqlServerTrustDefault({ ...conn });
      if (JSON.stringify(normalizedConn) !== JSON.stringify(conn)) {
        Object.assign(conn, normalizedConn);
        modified = true;
      }
      if (conn.password || conn.sshPassword) {
        try {
          await SaveCredential(
            conn.id,
            conn.password || "",
            conn.sshPassword || "",
          );
          conn.password = "";
          conn.sshPassword = "";
          modified = true;
        } catch (e) {
          console.error("Failed to migrate credentials to keyring", e);
        }
      }
      if (modified) needsSave = true;
    }

    if (needsSave) {
      savedConnections.value = connections;
    }
  };

  const loadCurrentSshHostKeyInfo = async () => {
    if (!canTrustCurrentSshHost.value || isLoadingHostKeyInfo.value) return;

    isLoadingHostKeyInfo.value = true;
    try {
      const result = await GetSSHHostKeyInfo(
        config.sshHost.trim(),
        Number.isInteger(config.sshPort) && config.sshPort > 0 ? config.sshPort : 22,
      );

      if (result.error) {
        setError(result.error);
        sshHostKeyInfo.value = null;
      } else {
        sshHostKeyInfo.value = result;
        if (!expectedSshFingerprint.value.trim() && pinnedSshFingerprint.value) {
          expectedSshFingerprint.value = pinnedSshFingerprint.value;
        }
      }
    } catch (e: any) {
      setError(e?.toString?.() || "Failed to load SSH host key info.");
      sshHostKeyInfo.value = null;
    } finally {
      isLoadingHostKeyInfo.value = false;
    }
  };

  const trustCurrentSshHost = async () => {
    if (!canTrustCurrentSshHost.value || isTrustingHost.value) return;
    if (!sshHostKeyInfo.value) {
      await loadCurrentSshHostKeyInfo();
      if (!sshHostKeyInfo.value) return;
    }
    if (isFingerprintMismatch.value) {
      setError("Fingerprint does not match expected value. Please verify before trusting.");
      return;
    }
    if (isPinnedFingerprintMismatch.value) {
      setError("Current fingerprint differs from previously trusted fingerprint for this host. Possible MITM or host key rotation.");
      return;
    }

    isTrustingHost.value = true;
    try {
      const result = await TrustSSHHost(
        config.sshHost.trim(),
        Number.isInteger(config.sshPort) && config.sshPort > 0 ? config.sshPort : 22,
      );

      if (result === "Success") {
        clearError();
        testSuccess.value = "SSH host trusted. Please retry Test Connection.";
        if (sshHostKeyInfo.value) {
          const nextEntry: SSHTrustAuditEntry = {
            host: config.sshHost.trim(),
            port: Number.isInteger(config.sshPort) && config.sshPort > 0 ? config.sshPort : 22,
            pattern: sshHostKeyInfo.value.pattern,
            keyType: sshHostKeyInfo.value.keyType,
            fingerprint: sshHostKeyInfo.value.fingerprint,
            trustedAt: new Date().toISOString(),
          };
          sshTrustAudit.value = [nextEntry, ...sshTrustAudit.value].slice(0, 20);
        }
      } else {
        setError(result || "Failed to trust SSH host.");
      }
    } catch (e: any) {
      setError(e?.toString?.() || "Failed to trust SSH host.");
    } finally {
      isTrustingHost.value = false;
    }
  };

  const acceptPinnedFingerprintRotation = () => {
    if (!sshHostKeyInfo.value || !isPinnedFingerprintMismatch.value) return;

    const reason = sshRotationReason.value.trim();
    if (reason.length < 8) {
      setError("Please provide a rotation reason (at least 8 characters).");
      return;
    }
    if (sshRotationConfirmText.value.trim().toUpperCase() !== "ROTATE") {
      setError("Type ROTATE to confirm host key rotation.");
      return;
    }

    const entry: SSHTrustAuditEntry = {
      host: config.sshHost.trim(),
      port: Number.isInteger(config.sshPort) && config.sshPort > 0 ? config.sshPort : 22,
      pattern: sshHostKeyInfo.value.pattern,
      keyType: sshHostKeyInfo.value.keyType,
      fingerprint: sshHostKeyInfo.value.fingerprint,
      trustedAt: new Date().toISOString(),
      rotatedFromFingerprint: pinnedSshFingerprint.value,
      rotationReason: reason,
    };

    sshTrustAudit.value = [entry, ...sshTrustAudit.value].slice(0, 200);
    expectedSshFingerprint.value = entry.fingerprint;
    sshRotationReason.value = "";
    sshRotationConfirmText.value = "";
    clearError();
    testSuccess.value = "Pinned fingerprint rotation accepted. You can trust this host key now.";
  };

  const copyCurrentSshFingerprint = async () => {
    const fingerprint = sshHostKeyInfo.value?.fingerprint || "";
    if (!fingerprint) return;

    try {
      if (navigator?.clipboard?.writeText) {
        await navigator.clipboard.writeText(fingerprint);
      } else {
        const textArea = document.createElement("textarea");
        textArea.value = fingerprint;
        textArea.style.position = "fixed";
        textArea.style.left = "-9999px";
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand("copy");
        document.body.removeChild(textArea);
      }
      testSuccess.value = "Fingerprint copied to clipboard.";
    } catch (e: any) {
      setError(e?.toString?.() || "Failed to copy fingerprint.");
    }
  };

  const clearSshTrustAudit = () => {
    sshTrustAudit.value = [];
    sshTrustAuditSearch.value = "";
    testSuccess.value = "SSH trust history cleared.";
  };

  const exportSshTrustAudit = async () => {
    if (!sshTrustAudit.value.length) {
      setError("No SSH trust history to export.");
      return;
    }

    try {
      const stamp = new Date().toISOString().replace(/[:.]/g, "-");
      const filePath = await SelectExportFile(`ssh-trust-audit-${stamp}.json`);
      if (!filePath) return;

      const payload = JSON.stringify(sshTrustAudit.value, null, 2);
      const writeError = await WriteTextFile(filePath, payload);
      if (writeError) {
        setError(writeError);
        return;
      }

      testSuccess.value = `SSH trust history exported to ${filePath}`;
    } catch (e: any) {
      setError(e?.toString?.() || "Failed to export SSH trust history.");
    }
  };

  const importSshTrustAudit = async () => {
    try {
      const filePath = await SelectImportFile();
      if (!filePath) return;

      const content = await ReadTextFile(filePath);
      const parsed = JSON.parse(content);
      if (!Array.isArray(parsed)) {
        setError("Invalid SSH trust history file format.");
        return;
      }

      const incoming: SSHTrustAuditEntry[] = parsed
        .map((item: any) => ({
          host: String(item?.host ?? ""),
          port: Number(item?.port ?? 22),
          pattern: String(item?.pattern ?? ""),
          keyType: String(item?.keyType ?? ""),
          fingerprint: String(item?.fingerprint ?? ""),
          trustedAt: String(item?.trustedAt ?? new Date().toISOString()),
        }))
        .filter((item) => item.pattern && item.fingerprint);

      if (!incoming.length) {
        setError("No valid SSH trust entries found in file.");
        return;
      }

      const byKey = new Map<string, SSHTrustAuditEntry>();
      const putEntry = (entry: SSHTrustAuditEntry) => {
        const key = `${entry.pattern}::${entry.fingerprint}`;
        const current = byKey.get(key);
        if (!current) {
          byKey.set(key, entry);
          return;
        }
        const currentTime = Date.parse(current.trustedAt);
        const nextTime = Date.parse(entry.trustedAt);
        if ((Number.isNaN(currentTime) ? 0 : currentTime) <= (Number.isNaN(nextTime) ? 0 : nextTime)) {
          byKey.set(key, entry);
        }
      };

      sshTrustAudit.value.forEach(putEntry);
      incoming.forEach(putEntry);

      sshTrustAudit.value = Array.from(byKey.values())
        .sort((a, b) => Date.parse(b.trustedAt) - Date.parse(a.trustedAt))
        .slice(0, 200);

      testSuccess.value = `Imported ${incoming.length} SSH trust entries.`;
    } catch (e: any) {
      setError(e?.toString?.() || "Failed to import SSH trust history.");
    }
  };

  return {
    config,
    resetConfig,
    error,
    testSuccess,
    isTrustingHost,
    isLoadingHostKeyInfo,
    sshHostKeyInfo,
    expectedSshFingerprint,
    sshRotationReason,
    sshRotationConfirmText,
    isFingerprintMatch,
    isFingerprintMismatch,
    isPinnedFingerprintMismatch,
    pinnedSshFingerprint,
    sshTrustAudit,
    sshTrustAuditSearch,
    filteredSshTrustAudit,
    canTrustCurrentSshHost,
    isLoading,
    isTesting,
    isQuickConnecting,
    savedConnections,
    connectionLabel,
    fieldErrors,
    handleSelectSqliteFile,
    cancelConnection,
    connect,
    testConnection,
    removeConnection,
    selectConnection,
    editConnection,
    migrateSavedConnections,
    loadCurrentSshHostKeyInfo,
    trustCurrentSshHost,
    acceptPinnedFingerprintRotation,
    copyCurrentSshFingerprint,
    clearSshTrustAudit,
    exportSshTrustAudit,
    importSshTrustAudit,
  };
}

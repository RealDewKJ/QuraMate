import { ref, reactive, computed } from "vue";
import { useLocalStorage, useToggle, watchImmediate } from "@vueuse/core";
import {
  ConnectDB,
  TestConnection,
  SetReadOnly,
  SelectSqliteFile,
  SaveCredential,
  DeleteCredential,
} from "../../wailsjs/go/main/App";

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
  sshEnabled: false,
  sshHost: "",
  sshPort: 22,
  sshUser: "",
  sshPassword: "",
  sshKeyFile: "",
  encoding: "",
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
) {
  const config = reactive<ConnectionConfig>({ ...DEFAULT_CONFIG });

  const resetConfig = () => {
    Object.assign(config, { ...DEFAULT_CONFIG });
  };

  const error = ref("");
  const testSuccess = ref("");
  const connectionToken = ref(0);
  const [isLoading] = useToggle(false);
  const [isTesting] = useToggle(false);
  const [isQuickConnecting] = useToggle(false);
  const savedConnections = useLocalStorage<ConnectionConfig[]>("savedConnections", []);

  const connectionLabel = computed(() => getConnectionLabel(config));

  // Auto-adjust port when db type changes
  watchImmediate(
    () => config.type,
    (newType) => {
      if (
        newType === "postgres" ||
        newType === "greenplum" ||
        newType === "redshift"
      )
        config.port = 5432;
      else if (newType === "cockroachdb") config.port = 26257;
      else if (newType === "mysql" || newType === "mariadb") config.port = 3306;
      else if (newType === "databend") config.port = 3307;
      else if (newType === "mssql") config.port = 1433;
    },
  );

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
    error.value = "Connection cancelled by user.";
    testSuccess.value = "";
  };

  const saveConnection = async (newConfig: ConnectionConfig) => {
    if (!newConfig.id) {
      const existing = savedConnections.value.find((c) =>
        isConfigEqual(c, newConfig),
      );
      if (existing && existing.id) {
        newConfig.id = existing.id;
      } else {
        newConfig.id = crypto.randomUUID();
      }
    }

    config.id = newConfig.id;

    if (newConfig.password || newConfig.sshPassword) {
      try {
        await SaveCredential(
          newConfig.id,
          newConfig.password || "",
          newConfig.sshPassword || "",
        );
      } catch (e) {
        console.error("Failed to save credentials to keyring", e);
      }
    }

    const storageConfig: ConnectionConfig = { ...newConfig, password: "", sshPassword: "" };

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
    error.value = "";
    testSuccess.value = "";
    isLoading.value = true;
    const currentToken = ++connectionToken.value;

    // Check for existing connection
    const existing = activeConnections().find((c) =>
      isConfigEqual(c.config, config),
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
      const result = await ConnectDB(config);
      if (currentToken !== connectionToken.value) return;

      if (result.id) {
        await saveConnection({ ...config });
        if (currentToken !== connectionToken.value) return;

        emitConnected({
          id: result.id,
          name: config.name || getConnectionLabel(config),
          config: { ...config },
        });
        resetConfig();
      } else {
        error.value = result.error || "Unknown error";
      }
    } catch (e: any) {
      if (currentToken !== connectionToken.value) return;
      error.value = e.toString();
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
    error.value = "";
    testSuccess.value = "";
    isTesting.value = true;
    const currentToken = ++connectionToken.value;

    try {
      const testConfig = { ...config };
      if (!testConfig.id) {
        testConfig.id = "";
      }
      const result = await TestConnection(testConfig);
      if (currentToken !== connectionToken.value) return;

      if (result === "Success") {
        testSuccess.value = "Connection successful!";
      } else {
        error.value = result;
      }
    } catch (e: any) {
      if (currentToken !== connectionToken.value) return;
      error.value = e.toString();
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
    config.name = conn.name || "";
    Object.assign(config, conn);
    config.password = "";
    config.sshPassword = "";

    isQuickConnecting.value = true;
    connect();
  };

  const editConnection = (conn: ConnectionConfig) => {
    config.name = conn.name || "";
    Object.assign(config, conn);
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

  return {
    config,
    resetConfig,
    error,
    testSuccess,
    isLoading,
    isTesting,
    isQuickConnecting,
    savedConnections,
    connectionLabel,
    handleSelectSqliteFile,
    cancelConnection,
    connect,
    testConnection,
    removeConnection,
    selectConnection,
    editConnection,
    migrateSavedConnections,
  };
}

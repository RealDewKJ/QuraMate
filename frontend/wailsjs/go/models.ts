export namespace main {
	
	export class BatchInsertResult {
	    inserted: number;
	    skipped: number;
	    errors: string[];
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchInsertResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inserted = source["inserted"];
	        this.skipped = source["skipped"];
	        this.errors = source["errors"];
	        this.error = source["error"];
	    }
	}
	export class ColumnDefinition {
	    name: string;
	    type: string;
	    nullable: boolean;
	    defaultValue: any;
	    primaryKey: boolean;
	    autoIncrement: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ColumnDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.nullable = source["nullable"];
	        this.defaultValue = source["defaultValue"];
	        this.primaryKey = source["primaryKey"];
	        this.autoIncrement = source["autoIncrement"];
	    }
	}
	export class ColumnChange {
	    oldName: string;
	    newDefinition: ColumnDefinition;
	
	    static createFrom(source: any = {}) {
	        return new ColumnChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldName = source["oldName"];
	        this.newDefinition = this.convertValues(source["newDefinition"], ColumnDefinition);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ConnectResult {
	    id: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.error = source["error"];
	    }
	}
	export class DBConfig {
	    id: string;
	    type: string;
	    host: string;
	    port: number;
	    user: string;
	    password: string;
	    database: string;
	    encoding?: string;
	    readOnly: boolean;
	    sshEnabled: boolean;
	    sshHost: string;
	    sshPort: number;
	    sshUser: string;
	    sshPassword: string;
	    sshKeyFile: string;
	
	    static createFrom(source: any = {}) {
	        return new DBConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.encoding = source["encoding"];
	        this.readOnly = source["readOnly"];
	        this.sshEnabled = source["sshEnabled"];
	        this.sshHost = source["sshHost"];
	        this.sshPort = source["sshPort"];
	        this.sshUser = source["sshUser"];
	        this.sshPassword = source["sshPassword"];
	        this.sshKeyFile = source["sshKeyFile"];
	    }
	}
	export class DatabaseInfo {
	    size: string;
	    tableCount: number;
	    viewCount: number;
	    routineCount: number;
	    dbName: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.size = source["size"];
	        this.tableCount = source["tableCount"];
	        this.viewCount = source["viewCount"];
	        this.routineCount = source["routineCount"];
	        this.dbName = source["dbName"];
	        this.version = source["version"];
	    }
	}
	export class ForeignKey {
	    table: string;
	    column: string;
	    refTable: string;
	    refColumn: string;
	    constraint: string;
	
	    static createFrom(source: any = {}) {
	        return new ForeignKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = source["table"];
	        this.column = source["column"];
	        this.refTable = source["refTable"];
	        this.refColumn = source["refColumn"];
	        this.constraint = source["constraint"];
	    }
	}
	export class IndexDefinition {
	    name: string;
	    columns: string[];
	    unique: boolean;
	    primary: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IndexDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.columns = source["columns"];
	        this.unique = source["unique"];
	        this.primary = source["primary"];
	    }
	}
	export class LogEntry {
	    time: string;
	    level: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.level = source["level"];
	        this.message = source["message"];
	    }
	}
	export class QueryHistoryEntry {
	    id: number;
	    query: string;
	    db_type: string;
	    timestamp: string;
	    is_favorite: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QueryHistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.query = source["query"];
	        this.db_type = source["db_type"];
	        this.timestamp = source["timestamp"];
	        this.is_favorite = source["is_favorite"];
	    }
	}
	export class ResultSet {
	    columns: string[];
	    rows: any[][];
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new ResultSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.message = source["message"];
	    }
	}
	export class QueryResult {
	    resultSets: ResultSet[];
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.resultSets = this.convertValues(source["resultSets"], ResultSet);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SSHHostKeyInfo {
	    host: string;
	    port: number;
	    pattern: string;
	    keyType: string;
	    fingerprint: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHHostKeyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.pattern = source["pattern"];
	        this.keyType = source["keyType"];
	        this.fingerprint = source["fingerprint"];
	        this.error = source["error"];
	    }
	}
	export class ServerProcess {
	    sessionId: string;
	    user: string;
	    host: string;
	    database: string;
	    command: string;
	    status: string;
	    state: string;
	    waitTime: number;
	    waitType: string;
	    queryText: string;
	    elapsedTime: number;
	    headBlock: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerProcess(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.user = source["user"];
	        this.host = source["host"];
	        this.database = source["database"];
	        this.command = source["command"];
	        this.status = source["status"];
	        this.state = source["state"];
	        this.waitTime = source["waitTime"];
	        this.waitType = source["waitType"];
	        this.queryText = source["queryText"];
	        this.elapsedTime = source["elapsedTime"];
	        this.headBlock = source["headBlock"];
	    }
	}
	export class TableChanges {
	    renameTable: string;
	    addColumns: ColumnDefinition[];
	    dropColumns: string[];
	    alterColumns: ColumnChange[];
	    addIndexes: IndexDefinition[];
	    dropIndexes: string[];
	    addFKs: ForeignKey[];
	    dropFKs: string[];
	
	    static createFrom(source: any = {}) {
	        return new TableChanges(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.renameTable = source["renameTable"];
	        this.addColumns = this.convertValues(source["addColumns"], ColumnDefinition);
	        this.dropColumns = source["dropColumns"];
	        this.alterColumns = this.convertValues(source["alterColumns"], ColumnChange);
	        this.addIndexes = this.convertValues(source["addIndexes"], IndexDefinition);
	        this.dropIndexes = source["dropIndexes"];
	        this.addFKs = this.convertValues(source["addFKs"], ForeignKey);
	        this.dropFKs = source["dropFKs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateInfo {
	    available: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    releaseNotes: string;
	    downloadURL: string;
	    publishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseNotes = source["releaseNotes"];
	        this.downloadURL = source["downloadURL"];
	        this.publishedAt = source["publishedAt"];
	    }
	}

}


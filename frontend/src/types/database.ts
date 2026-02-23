export interface ColumnMetadata {
    name: string;
    type: string;
    length: number;
    nullable: boolean;
}

export interface ResultSet {
    columns: string[];
    columnTypes?: ColumnMetadata[];
    rows: any[];
    message?: string;
}

export interface ForeignKey {
    table: string;
    column: string;
    refTable: string;
    refColumn: string;
    constraint: string;
}

export interface ColumnDefinition {
    name: string;
    type: string;
    length: number;
    nullable: boolean;
    primaryKey: boolean;
    autoIncrement: boolean;
    default?: string;
}

export interface IndexDefinition {
    name: string;
    columns: string[];
    unique: boolean;
    primary: boolean;
}

export type Transaction = any

export interface TransactionManager {
  execute<T>(callback: TransactionCallback<T>, opts?: Opts): Promise<T>;
}

export type TransactionCallback<T> = (transaction: Transaction) => Promise<T>;

export enum TransactionPropagation {
  CREATE_NEW,
  REUSE_EXISTENT
}

export type Opts = {
  transaction?: Transaction
  propagation?: TransactionPropagation
}

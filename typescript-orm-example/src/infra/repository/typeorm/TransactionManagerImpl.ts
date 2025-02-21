import { EntityManager } from "typeorm";
import { TransactionManager, TransactionCallback, Opts, TransactionPropagation, Transaction } from "../../../repositories/TransactionManager";
import { AppDataSource } from "../../datasource/datasource";

export class TypeOrmTransaction implements Transaction {
  constructor(private readonly transactionEntityManager: EntityManager) {}

  getEntityManager(): EntityManager {
    return this.transactionEntityManager;
  }
}

export class TransactionManagerImpl implements TransactionManager {

  async execute<T>(callback: TransactionCallback<T>, opts?: Opts): Promise<T> {

    switch (opts?.propagation) {
      case TransactionPropagation.REUSE_EXISTENT:
        const currentTx =  opts?.transaction;
        if (currentTx) {
          return callback(currentTx);
        } else {
          return this.runNewTransaction<T>(callback);
        }
      default:
        return this.runNewTransaction<T>(callback);
    }
  }

  private runNewTransaction<T>(callback: TransactionCallback<T>): Promise<T> {
    return AppDataSource.transaction(async (manager) => {
      const transaction = new TypeOrmTransaction(manager);
      return callback(transaction);
    });
  }
}

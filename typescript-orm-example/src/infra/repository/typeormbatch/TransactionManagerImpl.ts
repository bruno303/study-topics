import { EntityManager } from "typeorm/entity-manager/EntityManager";
import { Opts, Transaction, TransactionCallback, TransactionManager, TransactionPropagation } from "../../../repositories/TransactionManager";
import { AppDataSource } from "../../datasource/datasource";

type Operation = (em: EntityManager) => Promise<any>

export class TypeOrmTransaction implements Transaction {
  private readonly operations = new Array<Operation>();
  
  add(op: Operation) {
    this.operations.push(op);
  }

  getOperations(): Array<Operation> {
    return this.operations;
  }

  async commit() {
    await AppDataSource.transaction(async em => {
      for (const op of this.operations) {
        await op(em);
      }
    });
  }
}

export class TransactionManagerImpl implements TransactionManager {
  async execute<T>(callback: TransactionCallback<T>, opts?: Opts): Promise<T> {
    if (opts?.transaction && opts.propagation === TransactionPropagation.REUSE_EXISTENT) {
      return callback(opts.transaction)
    }
  
    const tx = new TypeOrmTransaction()
    const result = await callback(tx);
    await tx.commit();
    return result;
  }
}

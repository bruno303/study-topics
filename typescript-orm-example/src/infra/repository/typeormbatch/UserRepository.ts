import { User } from '../../../entities/User';
import { Opts } from '../../../repositories/TransactionManager';
import { AppDataSource } from '../../datasource/datasource';
import { TypeOrmTransaction } from './TransactionManagerImpl';

export class UserRepository {
  save(user: User, opts?: Opts): Promise<User> {
    const tx = this.getTransaction(opts);
    if (tx) {
      tx.add(em => em.save(user));
      return Promise.resolve(user);
    }
    return AppDataSource.getRepository(User).save(user);
  }

  getTransaction(opts?: Opts): TypeOrmTransaction | undefined {
    return (opts?.transaction as TypeOrmTransaction) || undefined;
  }
}

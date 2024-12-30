import { User } from '../entities/User';
import { TransactionManager, TransactionPropagation } from '../repositories/TransactionManager';
import { UserRepository } from '../repositories/UserRepository';

export class UserService {
  constructor(
    private readonly transactionManager: TransactionManager,
    private readonly userRepository: UserRepository
  ) {}

  async saveUser(name: string, email: string): Promise<User[]> {

    return await this.transactionManager.execute(async (tx1) => {
      return await this.transactionManager.execute(async (tx) => {
        const opts = { transaction: tx };

        const user = this.createUser(name, email);
        const user2 = this.createUser(email, name);
        const result = await Promise.all([
          this.userRepository.save(user, opts),
          this.userRepository.save(user2, opts)
        ])
        return result;
      }, { transaction: tx1, propagation: TransactionPropagation.REUSE_EXISTENT });

    }, { propagation: TransactionPropagation.CREATE_NEW });
  }

  private createUser(name: string, email: string) {
    const user = new User();
    user.email = email;
    user.name = name;
    return user;
  }
}


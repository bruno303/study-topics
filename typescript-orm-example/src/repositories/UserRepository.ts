import { User } from "../entities/User"
import { Opts } from "./TransactionManager"

export interface UserRepository {
  save(user: User, opts?: Opts): Promise<User>
}

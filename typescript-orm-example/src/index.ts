import 'reflect-metadata';
import express from 'express';
import { createConnection } from 'typeorm';
import { UserController } from './controllers/UserController';
import { UserService } from './services/UserService';
import { UserRepository } from './infra/repository/UserRepository';
import { TransactionManagerImpl } from './infra/repository/TransactionManagerImpl';

const app = express();
app.use(express.json());

// Initialize TypeORM
createConnection({
  type: 'sqlite',
  database: 'database.sqlite',
  synchronize: true,
  logging: false,
  entities: [__dirname + '/entities/*.{js,ts}']
}).then(() => {
  console.log('Database connected');

  const transactionManager = new TransactionManagerImpl();
  const userRepository = new UserRepository();
  const userService = new UserService(transactionManager, userRepository);
  const userController = new UserController(userService);

  // Set up routes
  app.post('/users', (req, res) => userController.createUser(req, res));

  app.listen(3000, () => {
    console.log('Server is running on http://localhost:3000');
  });
}).catch(error => console.log(error));

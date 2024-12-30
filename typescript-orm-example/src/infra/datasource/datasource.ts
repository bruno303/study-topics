import { DataSource, EntitySchema } from "typeorm";
import { User } from "../../entities/User";

export const AppDataSource = new DataSource({
    type: "sqlite",
    database: "test",
    synchronize: true,
    logging: "all",
    entities: [User],
    subscribers: [],
    migrations: [],
})

AppDataSource.initialize()
    .then(() => {
        console.log("Data Source has been initialized!")
    })
    .catch((err) => {
        console.error("Error during Data Source initialization", err)
    })

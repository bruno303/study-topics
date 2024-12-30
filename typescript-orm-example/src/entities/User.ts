import { Entity, PrimaryGeneratedColumn, Column } from 'typeorm';

@Entity({
  name: "users"
})
export class User {
  @PrimaryGeneratedColumn("uuid")
  string!: string;

  @Column()
  name!: string;

  @Column()
  email!: string;
}

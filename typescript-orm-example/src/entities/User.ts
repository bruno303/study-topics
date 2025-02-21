import { Entity, PrimaryGeneratedColumn, Column, Unique } from 'typeorm';

@Entity({
  name: "users"
})
@Unique(["email"])
export class User {
  @PrimaryGeneratedColumn("uuid")
  string!: string;

  @Column()
  name!: string;

  @Column()
  email!: string;
}

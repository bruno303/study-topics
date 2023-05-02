# Password-Hasher

## Conceitos
- Encoding é codificar a mensagem para um padrão que o receptor entenda. Temos encodings UTF-8, UTF-16, ASCII, etc. Encoding não protege as informações e é facilmente reversível.
- Hash são usados para validar integridade dos dados. São algoritmos unidirecionais, portanto não são reversíveis. Um potencial atacante teria muito trabalho para gerar várias entradas e combinações de hash até chegar em um hash salvo anteriormente, isso dificulta os ataques.
- Encrypt é tornar os dados ilegíveis por qualquer pessoa que não possua a chave usada na criptografia. A chave pode ser única ou um par de chaves (uma pública e uma privada).

## Referências
- [encoding-hashing-e-encryption-qual-a-diferenca](https://dev.to/kotlinautas/encoding-hashing-e-encryption-qual-a-diferenca-29gg)
- [hashing-passwords-in-java-with-bcrypt](https://dzone.com/articles/hashing-passwords-in-java-with-bcrypt)
- [kotlin-hashing](https://www.knowledgefactory.net/2021/01/kotlin-hashing.html)
- [java-password-hashing](https://www.baeldung.com/java-password-hashing)
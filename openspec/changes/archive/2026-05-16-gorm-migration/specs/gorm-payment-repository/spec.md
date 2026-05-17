## ADDED Requirements

### Requirement: Repositório GORM implementa PaymentRepository
O sistema SHALL fornecer uma implementação de `PaymentRepository` usando GORM como ORM, com um model interno separado de `domain.Payment` e conversão explícita entre os dois.

#### Scenario: Inicialização cria schema automaticamente
- **WHEN** `NewGORMRepository(dsn)` é chamado com uma DSN PostgreSQL válida
- **THEN** o GORM executa `AutoMigrate` e garante que a tabela `payments` existe com todas as colunas esperadas

#### Scenario: Falha de conexão retorna erro
- **WHEN** `NewGORMRepository(dsn)` é chamado com uma DSN inválida ou banco inacessível
- **THEN** a função retorna um erro descritivo e nil como repositório

---

### Requirement: Save persiste um Payment novo
O repositório SHALL inserir um novo `domain.Payment` na tabela `payments`, convertendo-o para o model GORM antes da inserção.

#### Scenario: Inserção bem-sucedida
- **WHEN** `Save(ctx, payment)` é chamado com um `domain.Payment` válido
- **THEN** o registro é inserido na tabela `payments` e nenhum erro é retornado

#### Scenario: Inserção com ID duplicado retorna erro
- **WHEN** `Save(ctx, payment)` é chamado com um `payment.ID` que já existe na tabela
- **THEN** um erro de constraint é retornado

---

### Requirement: FindByOrderID localiza pagamento pelo order_id
O repositório SHALL retornar o `domain.Payment` correspondente ao `order_id` informado.

#### Scenario: Registro encontrado
- **WHEN** `FindByOrderID(ctx, orderID)` é chamado com um `order_id` existente
- **THEN** retorna o `domain.Payment` correspondente sem erro

#### Scenario: Registro não encontrado
- **WHEN** `FindByOrderID(ctx, orderID)` é chamado com um `order_id` inexistente
- **THEN** retorna erro com mensagem "registro não encontrado"

---

### Requirement: FindByPaymentID localiza pagamento pelo payment_id
O repositório SHALL retornar o `domain.Payment` correspondente ao `payment_id` informado.

#### Scenario: Registro encontrado
- **WHEN** `FindByPaymentID(ctx, paymentID)` é chamado com um `payment_id` existente
- **THEN** retorna o `domain.Payment` correspondente sem erro

#### Scenario: Registro não encontrado
- **WHEN** `FindByPaymentID(ctx, paymentID)` é chamado com um `payment_id` inexistente
- **THEN** retorna erro com mensagem "registro não encontrado"

---

### Requirement: UpdatePayment atualiza status do pagamento
O repositório SHALL atualizar `payment_id`, `net_amount`, `status`, `business_status`, `saga_status` e `updated_at` do registro identificado por `order_id`.

#### Scenario: Atualização bem-sucedida
- **WHEN** `UpdatePayment(ctx, orderID, ...)` é chamado com um `order_id` existente
- **THEN** os campos são atualizados na tabela e nenhum erro é retornado

#### Scenario: order_id não encontrado
- **WHEN** `UpdatePayment(ctx, orderID, ...)` é chamado com um `order_id` inexistente
- **THEN** retorna erro informando que o pagamento não foi encontrado

---

### Requirement: Model GORM mapeia colunas nullable com ponteiros
O model interno `paymentModel` SHALL usar `*string` para `payment_id` e `*float64` para `net_amount` e `*time.Time` para `expires_at`, permitindo que GORM escreva `NULL` nativamente sem helpers manuais.

#### Scenario: Campo nullable nil salvo como NULL
- **WHEN** `domain.Payment.PaymentID` está vazio e `Save` é chamado
- **THEN** a coluna `payment_id` é gravada como `NULL` no banco

#### Scenario: Campo nullable preenchido salvo com valor
- **WHEN** `domain.Payment.PaymentID` tem valor e `Save` é chamado
- **THEN** a coluna `payment_id` é gravada com o valor correspondente

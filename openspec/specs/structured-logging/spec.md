### Requirement: Logger JSON global inicializado em main.go
A aplicação SHALL inicializar um logger JSON estruturado via `slog.SetDefault` com `slog.NewJSONHandler(os.Stdout, nil)` como primeira ação em `main()`, antes de qualquer outro log, para que todos os pacotes usem automaticamente o formato JSON.

#### Scenario: Output em JSON lines
- **WHEN** a aplicação gera qualquer log em qualquer nível
- **THEN** o output SHALL ser uma linha JSON válida com campos `time`, `level` e `msg` no mínimo

#### Scenario: Logger disponível globalmente sem injeção
- **WHEN** qualquer pacote chama `slog.Info`, `slog.Warn` ou `slog.Error`
- **THEN** o logger SHALL usar o JSONHandler configurado em `main.go` sem que o logger seja passado por parâmetro

---

### Requirement: PaymentService loga entrada e saída de ProcessPaymentRequest
O método `ProcessPaymentRequest` SHALL registrar log estruturado no início (campos do evento recebido) e ao finalizar (resultado e duração), usando `slog.Info` para sucesso e `slog.Error` para falha.

#### Scenario: Log de início com campos contextuais
- **WHEN** `ProcessPaymentRequest` é invocado
- **THEN** o sistema SHALL logar `op=ProcessPaymentRequest`, `order_id`, `correlation_id` e `event_type`

#### Scenario: Log de conclusão com duração
- **WHEN** `ProcessPaymentRequest` conclui com sucesso
- **THEN** o sistema SHALL logar `preference_id`, `checkout_url` e `duration_ms`

#### Scenario: Log de erro com causa
- **WHEN** `ProcessPaymentRequest` retorna erro em qualquer etapa
- **THEN** o sistema SHALL logar `error` com a mensagem de erro e `duration_ms`

---

### Requirement: PaymentService loga entrada e saída de ProcessWebhook
O método `ProcessWebhook` SHALL registrar log estruturado no início e ao finalizar, incluindo `payment_id`, status consultado no MP e resultado.

#### Scenario: Log de início
- **WHEN** `ProcessWebhook` é invocado
- **THEN** o sistema SHALL logar `op=ProcessWebhook` e `payment_id`

#### Scenario: Log de conclusão com status
- **WHEN** `ProcessWebhook` conclui
- **THEN** o sistema SHALL logar `mp_status`, `business_status` resultante e `duration_ms`

---

### Requirement: MercadoPago Client loga cada chamada à API com duração
O cliente MercadoPago SHALL registrar log estruturado antes e após cada chamada à API do MP (`CreatePreference`, `GetPaymentStatus`), incluindo a duração da chamada HTTP.

#### Scenario: Log de CreatePreference
- **WHEN** `CreatePreference` é chamado
- **THEN** o sistema SHALL logar `op=CreatePreference`, `order_id` e ao concluir `preference_id` e `duration_ms`

#### Scenario: Log de GetPaymentStatus
- **WHEN** `GetPaymentStatus` é chamado
- **THEN** o sistema SHALL logar `op=GetPaymentStatus`, `payment_id` e ao concluir `mp_status` e `duration_ms`

#### Scenario: Log de erro na API do MP
- **WHEN** a API do MP retorna erro
- **THEN** o sistema SHALL logar `error` com a resposta do MP e `duration_ms`

---

### Requirement: SNS Publisher loga cada publicação com event_type e duração
O publisher SNS SHALL registrar log estruturado ao publicar cada evento, com `event_type`, `order_id`, `correlation_id` e duração da chamada ao AWS SNS.

#### Scenario: Log de publicação bem-sucedida
- **WHEN** um evento é publicado com sucesso no SNS
- **THEN** o sistema SHALL logar `event_type`, `order_id`, `correlation_id` e `duration_ms`

#### Scenario: Log de falha na publicação
- **WHEN** a publicação no SNS retorna erro
- **THEN** o sistema SHALL logar `error` e `duration_ms` em nível ERROR

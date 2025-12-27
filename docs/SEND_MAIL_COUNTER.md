# Funcionalidade: Envio de Contadores por Email

## Comando: `/send_mail_counter`

Este comando permite enviar um relatório dos contadores de impressoras por email, incluindo uma planilha Excel anexada.

### Uso

```
/send_mail_counter <email1> [email2] [email3] ...
```

### Exemplos

```bash
# Enviar para um único destinatário
/send_mail_counter joao@empresa.com

# Enviar para múltiplos destinatários
/send_mail_counter joao@empresa.com maria@empresa.com ti@empresa.com
```

### Configuração

Para usar esta funcionalidade, você precisa configurar as seguintes variáveis de ambiente no arquivo `.env`:

```env
SMTP_SERVER=smtp.gmail.com:587
SMTP_USER=seu-email@gmail.com
SMTP_PASSWORD=sua-senha-de-app
```

#### Configuração para Gmail

1. Ative a verificação em duas etapas na sua conta Google
2. Gere uma senha de app em: <https://myaccount.google.com/apppasswords>
3. Use essa senha no campo `SMTP_PASSWORD`

#### Outros provedores SMTP

- **Outlook/Hotmail**: `smtp-mail.outlook.com:587`
- **Yahoo**: `smtp.mail.yahoo.com:587`
- **Office 365**: `smtp.office365.com:587`

### O que o comando faz

1. ✅ Coleta os contadores de todas as impressoras do Zabbix
2. ✅ Gera uma planilha Excel com os dados
3. ✅ Cria um email HTML formatado com uma tabela dos contadores
4. ✅ Anexa a planilha Excel ao email
5. ✅ Envia para todos os destinatários especificados

### Formato do Email

O email enviado contém:

- **Assunto**: "Relatório de Contadores de Impressoras"
- **Corpo**: Tabela HTML formatada com os contadores
- **Anexo**: Planilha Excel com dados detalhados

### Refatoração do Mailer

O pacote `mailer` foi completamente refatorado para ser mais simples e robusto:

**Antes:**

```go
mailer := &Mailer{}
msg, _ := mailer.GetNewMail(from, to, subject)
mailer.SetHTMLBody(msg, htmlContent)
mailer.SendMail(envMap, msg)
```

**Depois:**

```go
client := mailer.NewClient() // Lê automaticamente do .env
err := client.SendEmail(mailer.EmailMessage{
    From:        "sender@example.com",
    To:          []string{"recipient@example.com"},
    Subject:     "Assunto",
    HTMLBody:    "<h1>HTML Content</h1>",
    Attachments: []string{"/path/to/file.xlsx"},
})
```

### Vantagens da refatoração

- ✅ Mais simples de usar
- ✅ Suporte nativo a múltiplos destinatários
- ✅ Suporte a anexos
- ✅ Melhor tratamento de erros
- ✅ Configuração automática via variáveis de ambiente
- ✅ Código mais limpo e testável

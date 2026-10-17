# Comando `/start`

## Descrição

Comando inicial do bot que exibe mensagem de boas-vindas para usuários autorizados ou mensagem de acesso negado para não autorizados.

## Funcionalidade

### Para Usuários Autorizados ✅

Quando um usuário autorizado envia `/start`, recebe uma mensagem de boas-vindas completa com:

- Saudação personalizada com o nome do usuário
- Apresentação do bot
- Menu organizado com todos os comandos disponíveis
- Dicas de uso

**Exemplo de mensagem:**

```
👋 Bem-vindo, João Silva!

Sou o LapaTelegramBot, seu assistente de gerenciamento e monitoramento.

🎯 Principais Funcionalidades:

🌐 Monitoramento de Rede
• /ping - Testa conectividade
• /listip - Lista hosts do Zabbix

📊 Monitoramento Zabbix
• /status_check - Status dos hosts
• /printers_counter - Contadores de impressoras
• /protheus_status - Status Protheus/TOTVS

⚙️ Gerenciamento de Serviços
• /services - Gerenciar serviços remotos
• /list_services - Listar serviços

💻 Gerenciamento Windows
• /restart_win - Reiniciar host
• /shutdown_win - Desligar host

🔐 Active Directory
• /resetpwd <usuario> - Redefinir senha, exigir troca no próximo logon e desbloquear conta

📧 Relatórios
• /send_mail_counter - Enviar contadores por email

⏰ Agendamentos
• /schedule_add - Criar agendamento
• /schedule_list - Listar agendamentos
• /schedule_remove - Remover agendamento
• /schedule_help - Ajuda sobre CRON

💡 Dica: Todos os comandos fornecem feedback em tempo real!

Digite qualquer comando para começar. 🚀
```

### Para Usuários Não Autorizados 🚫

Quando um usuário não autorizado tenta usar o bot, recebe uma mensagem clara de acesso negado:

**Exemplo de mensagem:**

```
🚫 Acesso Não Autorizado

Olá, Maria Santos!

Infelizmente você não tem permissão para usar este bot.

Este é um bot privado e apenas usuários autorizados podem utilizá-lo.

Chat ID: 123456789
```

## Segurança

### Registro de Tentativas

Todas as tentativas de acesso não autorizado são registradas no log do sistema:

```
⚠️  Tentativa de acesso não autorizado - Chat ID: 123456789, Nome: Maria Santos
```

### Comportamento

- ✅ Comando `/start` é processado **antes** da verificação de autorização
- ✅ Usuários não autorizados recebem feedback mas não podem usar outros comandos
- ✅ Chat ID é exibido para facilitar autorização posterior se necessário
- ✅ Logs ajudam a monitorar tentativas de acesso

## Vantagens

### 1. Experiência do Usuário

- Mensagem personalizada com nome do usuário
- Menu completo e organizado
- Fácil navegação pelos comandos

### 2. Segurança

- Feedback claro sobre acesso negado
- Registro de tentativas não autorizadas
- Chat ID disponível para autorização

### 3. Documentação Integrada

- Usuário vê todos os comandos disponíveis
- Descrições curtas e objetivas
- Organização por categorias

## Como Autorizar Novos Usuários

Se alguém tentar usar o bot e você quiser autorizá-lo:

1. Verifique o log para encontrar o Chat ID
2. Adicione o Chat ID ao `.env`:

   ```env
   TELEGRAM_ALLOWED_CHAT_ID=123,456,789456123
   ```

3. Reinicie o bot
4. Peça para o usuário enviar `/start` novamente

## Integração com BotFather

O comando `/start` está configurado no BotFather como:

```
start - Inicia o bot e exibe menu de comandos
```

Este é o primeiro comando da lista, facilitando a descoberta pelos usuários.

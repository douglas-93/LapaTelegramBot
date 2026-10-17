# TelegramNotify

Este projeto foi criado inicialmente para treinar Go, integrando o Zabbix com a API de Bots do Telegram, apenas para verificar os dispositivos online. Porém, diante da praticidade para mim, foram incrementadas novas funcionalidades.

Por se tratar de algo específico para mim, provavelmente as consultas não funcionarão para todos, sendo necessários ajustes no código, visto que não generalizei muito as consultas.

> ⚠️ **Nota**: Este é um projeto pessoal e não foi desenvolvido para ser aplicado de forma genérica. Você precisará adaptar as consultas ao Zabbix, IDs de grupos, nomes de items e outras configurações específicas do seu ambiente.

Mas, caso alguém queira tentar, você pode começar fazendo o seguinte.

## 🔧 Passos Iniciais

Utilize o `.env.example` para definir as variáveis abaixo:

```dotenv
TELEGRAM_API_TOKEN=<YOUR_TELEGRAM_BOT_TOKEN>
TELEGRAM_ALLOWED_CHAT_ID=123,456
ZABBIX_API_TOKEN=<YOUR_ZABBIX_TOKEN>
ZABBIX_API_URL=<YOUR_ZABBIX_SERVER_ADDRESS>/zabbix/api_jsonrpc.php
SMTP_SERVER=smtp.gmail.com:587
SMTP_USER=seu-email@gmail.com
SMTP_PASSWORD=sua-senha-de-app
AD_DEFAULT_PASSWORD=
```

### Configuração do Telegram

Para o Telegram, você precisará chamar o [@BotFather](https://t.me/botfather) e criar uma chave. Como o Telegram não disponibiliza uma ferramenta de visibilidade do bot, foi necessário fazer a validação via código, onde eu capturo o ChatID (aparecerá no log assim que seu bot for acionado) e defino que ele está autorizado a conversar com esse Chat.

#### <span style="color:red">⚠️ ATENÇÃO!</span>

Cuide bem da sua chave, pois, qualquer um com acesso a ela, terá controle total de seu Bot.

### Configuração do Zabbix

Para o Zabbix, existem duas alternativas: você capturar o token via autenticação, ou definir um token já no Zabbix. Por praticidade e facilidade de revogação caso necessário, optei pela segunda opção.

Será necessário definir o TOKEN e o endereço do seu servidor Zabbix, lembrando de manter o `/api_jsonrpc.php` que é o ponto de chamada da API.

**Documentação oficial:**

- [Telegram Bot API](https://core.telegram.org/bots/tutorial#introduction)
- [Zabbix API](https://www.zabbix.com/documentation/current/en/manual/api)

### Configuração SMTP (Opcional)

Para usar o comando `/send_mail_counter`, configure as variáveis SMTP:

- **SMTP_SERVER**: Servidor SMTP e porta (ex: `smtp.gmail.com:587`)
- **SMTP_USER**: Email de envio
- **SMTP_PASSWORD**: Senha do email ou senha de app

#### Configuração para Gmail

1. Ative a verificação em duas etapas
2. Gere uma senha de app em: [https://myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
3. Use essa senha no campo `SMTP_PASSWORD`

#### Outros provedores

- **Outlook/Hotmail**: `smtp-mail.outlook.com:587`
- **Yahoo**: `smtp.mail.yahoo.com:587`
- **Office 365**: `smtp.office365.com:587`

### Redefinição de senha no Active Directory

Configure `AD_DEFAULT_PASSWORD` com a senha padrão que será aplicada pelo comando `/resetpwd <usuario>`. O bot deve rodar em um computador Windows com o módulo PowerShell ActiveDirectory disponível e sob uma conta com permissão para redefinir senhas e desbloquear usuários no domínio.

O comando redefine a senha, exige que o usuário a troque no próximo logon e desbloqueia a conta somente se ela estiver bloqueada. Se a conta estiver configurada com senha que nunca expira, essa opção será desativada para que a troca obrigatória possa ser aplicada. Se o módulo não estiver disponível, a conta não for encontrada ou a conta de execução não tiver permissão, o bot informa o erro. Proteja o arquivo `.env`, pois ele contém a senha padrão.

### ⚙️ Ajustes Necessários

Como este projeto foi desenvolvido para um ambiente específico, você precisará ajustar:

- **IDs de grupos no Zabbix**: No arquivo `printers.go`, o grupo de impressoras está hardcoded como `"22"`. Ajuste para o ID do seu grupo.
- **Keys de items**: Os items buscados (como `"icmpping"`, `"contador.colorido"`, `"TOTVS"`) precisam existir no seu Zabbix com os mesmos nomes, ou você deve alterar o código.
- **Comandos Windows**: Os comandos de restart/shutdown funcionam apenas em ambientes Windows com permissões adequadas.

## 🚀 Funcionalidades

### 🌐 Monitoramento de Rede

#### `/ping <ip1> <ip2> ...`

Realiza ping em um ou mais endereços IP simultaneamente.

- Mostra latência média, pacotes enviados/recebidos e taxa de perda
- Suporta múltiplos IPs em uma única execução
- Feedback em tempo real
- Exemplo: `/ping 192.168.0.1 192.168.0.2`

#### `/listip`

Lista todos os hosts e seus endereços IP cadastrados no Zabbix.

- Exibe hostname e interface(s) de rede
- Apenas hosts ativos são listados
- Feedback em tempo real

### 📊 Monitoramento Zabbix

#### `/status_check`

Verifica o status online/offline de todos os hosts monitorados.

- ✅ Host online (icmpping = 1)
- ❌ Host offline (icmpping = 0)
- Consulta em tempo real via API Zabbix
- Feedback progressivo

#### `/printers_counter`

Exibe os contadores de impressão das impressoras monitoradas.

- Contador preto e branco
- Contador colorido
- Contador total
- Gera planilha Excel formatada automaticamente
- Feedback multi-etapa (coleta → processamento → planilha)
- Apenas impressoras do grupo específico no Zabbix (ID: 22)

#### `/protheus_status`

Monitora o status dos serviços Protheus/TOTVS.

- ✅ Serviço rodando
- ❌ Serviço parado
- Consulta itens com key "TOTVS" no Zabbix
- Feedback em tempo real

### 💻 Gerenciamento de Hosts Windows

#### `/restart_win <hostname>`

Reinicia remotamente um host Windows.

- Requer permissões administrativas
- Exemplo: `/restart_win LVMAQUINA`

#### `/shutdown_win <hostname>`

Desliga remotamente um host Windows.

- Requer permissões administrativas
- Exemplo: `/shutdown_win LVMAQUINA`

#### `/resetpwd <usuario>`

Redefine a senha do usuário do Active Directory para o valor configurado em `AD_DEFAULT_PASSWORD`, exige a troca no próximo logon e desbloqueia a conta caso esteja bloqueada. Se a conta estiver com senha que nunca expira, essa configuração será desativada.

- Requer o módulo PowerShell ActiveDirectory e permissões adequadas para a conta que executa o bot
- Exemplo: `/resetpwd usuario`

### ⚙️ Gerenciamento de Serviços Remotos

#### `/services <host> <start|stop|restart> <serviço1> [serviço2] ...`

Gerencia serviços Windows em hosts remotos.

- **Ações**: `start`, `stop`, `restart`
- Suporta múltiplos serviços simultaneamente
- Execução concorrente para melhor performance
- Relatório detalhado de cada operação
- Feedback em tempo real
- Exemplos:
  - `/services 192.168.100.16 restart Spooler`
  - `/services SERVER01 stop wuauserv BITS`
  - `/services 192.168.1.10 start TOTVS_AppServer TOTVS_DBAccess`

#### `/list_services <host> [filtro]`

Lista todos os serviços de um host Windows remoto.

- Lista até 50 serviços por vez
- Filtro opcional por nome (case-insensitive)
- Feedback em tempo real
- Exemplos:
  - `/list_services 192.168.100.16` (lista todos)
  - `/list_services SERVER01 TOTVS` (filtra por "TOTVS")
  - `/list_services 192.168.1.10 SQL` (filtra por "SQL")

### 📧 Envio de Relatórios por Email

#### `/send_mail_counter <email1> [email2] ...`

Envia relatório de contadores de impressoras por email.

- Email HTML formatado com tabela
- Planilha Excel anexada
- Suporta múltiplos destinatários
- Feedback multi-etapa (coleta → planilha → envio)
- Requer configuração SMTP no `.env`
- Exemplos:
  - `/send_mail_counter joao@empresa.com`
  - `/send_mail_counter joao@empresa.com maria@empresa.com ti@empresa.com`

### 📁 Upload de Arquivos

#### Envio Direto

O bot agora aceita o envio de arquivos diretamente no chat.

- **Tipos suportados**: Documentos, Fotos, Áudios, Vídeos e Mensagens de Voz.
- **Local de salvamento**: Todos os arquivos são salvos na pasta `uploaded_files/` no diretório raiz do bot.
- **Feedback**: O bot informa quando o download inicia e quando é concluído com sucesso.

### ⏰ Sistema de Agendamento

#### `/schedule_add <min> <hora> <dia> <mês> <dia_semana> <comando>`

Cria um novo agendamento usando expressões CRON.

- Executa comandos automaticamente no horário especificado
- Suporta todos os comandos do bot
- Exemplo: `/schedule_add 0 8 20 * * printers_counter`
  - Executa `/printers_counter` todo dia 20 às 08:00

#### `/schedule_list`

Lista todos os agendamentos ativos.

- Mostra ID, expressão CRON e comando de cada agendamento
- IDs são necessários para remover agendamentos

#### `/schedule_remove <ID>`

Remove um agendamento específico pelo ID.

- Exemplo: `/schedule_remove 1764686892095287000`

#### `/schedule_help`

Exibe guia completo sobre expressões CRON.

- Exemplos práticos de agendamentos
- Formato: `MIN HORA DIA MÊS DIA_SEMANA`

## 📦 Dependências

```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get github.com/joho/godotenv
go get github.com/go-ping/ping
go get github.com/go-co-op/gocron
go get github.com/xuri/excelize/v2
go get github.com/wneessen/go-mail
go get golang.org/x/sys/windows
```

### Desenvolvido com o apoio destes pacotes

```text
github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
github.com/go-co-op/gocron v1.37.0
github.com/joho/godotenv v1.5.1
github.com/go-ping/ping v1.2.0
github.com/xuri/excelize/v2 v2.10.0
github.com/wneessen/go-mail v0.7.2
golang.org/x/sys v0.37.0
```

## 🚀 Execução

```bash
go run main.go
```

## 🔐 Segurança

- Apenas chat IDs autorizados podem usar o bot
- Comandos Windows requerem privilégios administrativos
- Agendamentos são persistidos em `schedules.json`
- Nunca compartilhe seu token do Telegram ou do Zabbix

## 📝 Logs

Todas as interações são registradas no console com:

- Chat ID do usuário
- Nome e username
- Comando executado
- Timestamp

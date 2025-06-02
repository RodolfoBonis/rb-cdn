# RB CDN

Um serviço robusto de CDN (Content Delivery Network) desenvolvido em Go para gerenciamento e distribuição de arquivos de mídia.

## 🚀 Sobre o Projeto

O RB CDN é um serviço especializado para upload e distribuição de arquivos de mídia, utilizando o MinIO como backend de armazenamento. O projeto foi desenvolvido com foco em performance, segurança e escalabilidade.

## ✨ Funcionalidades

- Upload de arquivos de mídia
- Distribuição de conteúdo via CDN
- Integração com MinIO para armazenamento
- Monitoramento com New Relic
- Logging com Sentry
- Documentação Swagger
- Métricas Prometheus
- Integração com RabbitMQ

## 🛠️ Tecnologias Utilizadas

- Go 1.20
- Gin Web Framework
- MinIO
- New Relic
- Sentry
- Prometheus
- RabbitMQ
- Swagger
- Docker

## 📋 Pré-requisitos

- Go 1.20 ou superior
- Docker e Docker Compose
- MinIO (configurado via Docker)
- New Relic Account (para monitoramento)
- Sentry Account (para logging)

## 🔧 Configuração do Ambiente

1. Clone o repositório:
```bash
git clone https://github.com/RodolfoBonis/rb-cdn.git
cd rb-cdn
```

2. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

3. Execute com Docker Compose:
```bash
docker-compose up -d
```

## 🚀 Executando o Projeto

### Localmente

```bash
go mod download
go run main.go
```

### Com Docker

```bash
docker build -t rb-cdn .
docker run -p 8080:8080 rb-cdn
```

## 📚 Documentação

A documentação da API está disponível via Swagger em:
```
https://rb-cdn.rodolfodebonis.com.br/swagger/index.html
```

## 🧪 Testes

Execute os testes com:
```bash
go test ./...
```

## 📦 Estrutura do Projeto

```
.
├── core/           # Núcleo da aplicação
├── routes/         # Definição de rotas
├── docs/           # Documentação
├── features/       # Funcionalidades
├── main.go         # Ponto de entrada
└── dockerfile      # Configuração Docker
```

## 🔒 Segurança

- CORS configurado
- Autenticação via API Key
- Headers de segurança
- Timeouts configurados
- Proxies confiáveis

## 📊 Monitoramento

- New Relic para APM
- Sentry para logging de erros
- Prometheus para métricas
- Logs estruturados com Zap

## 🤝 Contribuindo

1. Fork o projeto
2. Crie sua Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a Branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 👨‍💻 Autor

Rodolfo Bonis

## 📞 Suporte

Para suporte, envie um email para [seu-email@exemplo.com] ou abra uma issue no GitHub.

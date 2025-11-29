# Deployment Guide

## GitHub Secrets Configuration

Para que o build automático funcione, você precisa configurar os seguintes secrets no GitHub:

### 1. AWS Credentials (para ECR)

```
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
```

Obtenha as credenciais AWS que têm permissão para fazer push no ECR.

### 2. SSH Private Key (para repositórios privados)

```
SSH_PRIVATE_KEY
```

Esta é a chave SSH privada que tem acesso ao repositório `rb_auth_client`.

#### Como gerar e configurar a SSH Key:

1. Gere uma nova chave SSH (se ainda não tiver):
```bash
ssh-keygen -t ed25519 -C "github-actions-rb-cdn" -f ~/.ssh/github_actions_rb_cdn
```

2. Adicione a chave pública ao GitHub:
   - Vá em: https://github.com/settings/keys
   - Clique em "New SSH key"
   - Cole o conteúdo de `~/.ssh/github_actions_rb_cdn.pub`
   - Dê um nome como "GitHub Actions - rb-cdn"

3. Adicione a chave privada como secret:
   - Vá em: https://github.com/RodolfoBonis/rb-cdn/settings/secrets/actions
   - Clique em "New repository secret"
   - Nome: `SSH_PRIVATE_KEY`
   - Valor: Cole o conteúdo completo de `~/.ssh/github_actions_rb_cdn` (a chave PRIVADA)

### Como adicionar secrets no GitHub:

1. Acesse: https://github.com/RodolfoBonis/rb-cdn/settings/secrets/actions
2. Clique em "New repository secret"
3. Adicione cada secret com o nome e valor correspondente

## Build Manual Local

Para fazer build local da imagem:

```bash
# Certifique-se de que seu SSH agent está rodando
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519  # ou sua chave SSH

# Build da imagem
docker buildx build --ssh default -t rb-cdn:local .
```

## Versioning

O workflow usa tags git para versionar as imagens:

- Se existir uma tag (ex: v0.6.0), usa essa versão
- Se não existir tag, usa "0.6.0" como fallback
- Sempre adiciona o commit SHA para rastreabilidade
- Sempre marca como "latest" também

### Criar uma nova versão:

```bash
git tag v0.6.0
git push origin v0.6.0
```

Isso vai disparar o workflow e criar as imagens:
- `0.6.0`
- `0.6.0-abc1234` (com commit SHA)
- `abc1234` (apenas commit SHA)
- `latest`

## Atualizando o Deployment no Kubernetes

Após o build, atualize o arquivo de deployment:

```bash
# Em k3s-apps/applications/rb-cdn/rb-cdn-deployment.yaml
# Atualize a linha:
image: 718446585908.dkr.ecr.sa-east-1.amazonaws.com/rodolfobonis/rb-cdn:0.6.0-abc1234
```

Depois faça commit e push para aplicar via ArgoCD.

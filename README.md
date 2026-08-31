# 🎓 Sistema Acadêmico Multiplataforma

Sistema de cadastro e consulta de notas de alunos, composto por front-end responsivo,
API REST, banco de dados relacional e camada de cache — containerizado com Docker e
implantado em Kubernetes.

Exercício 07 da disciplina de Programação Multiplataforma — FATEC Antonio Russo.

---

## 👥 Equipe

- Carolina Pichelli Souza
- Fernando Alcantara D'Avila
- Guilherme Xavier Zanetti
- Heloísa Pichelli Souza
- Lucas Batista de Sousa
- Nuno Kasuo Tronco Yokoji

---

## ⚡ Atalho: Taskfile

O projeto traz um [`Taskfile.yml`](Taskfile.yml) que embrulha todos os comandos deste README.
Instale o Docker e o [Task](https://taskfile.dev) conforme a seção
**[Pré-requisitos](#-pré-requisitos)** logo abaixo — o resto o próprio Taskfile resolve.

```bash
task                  # lista todas as tarefas disponíveis
task tools            # instala kind e kubectl em ~/.local/bin (sem sudo)

task up               # sobe tudo no Docker Compose      → http://localhost:8080
task k8s:up           # sobe tudo no Kubernetes (kind)   → http://localhost:8080

task test             # testes de API, validação, Redis e Docker
task test:all         # tudo acima + Kubernetes, escalabilidade e persistência

task stop             # para tudo sem apagar nada
task clean            # apaga containers, volume, cluster e imagens
```

O Taskfile já cuida do `PATH` do kind/kubectl, libera a porta 8080 quando o outro ambiente
está no ar e carrega as imagens no cluster. **Compose e Kubernetes disputam a porta 8080**,
então só um roda por vez — `task up` e `task k8s:up` fazem essa troca sozinhos.

As seções abaixo trazem os comandos equivalentes, um a um, para quem preferir executá-los
manualmente ou precisar demonstrá-los na entrega.

---

## 🧰 Pré-requisitos

| Ferramenta | Para quê | Obrigatória? |
|---|---|---|
| **Docker Engine** + plugin Compose | construir as imagens e rodar tudo | sim |
| **Task** | atalhos do `Taskfile.yml` | não (há o comando manual equivalente para tudo) |
| **kind** | cluster Kubernetes local | só para a parte de Kubernetes |
| **kubectl** | operar o cluster | só para a parte de Kubernetes |
| **Go 1.26+** | rodar o back-end fora do container | não |

Nada mais precisa ser instalado: as dependências Go são baixadas durante o build da imagem, e
o front-end é HTML/CSS/JavaScript puro — sem Node, sem npm, sem etapa de build.

### 1. Docker Engine e plugin Compose

No Ubuntu/Debian/Linux Mint:

```bash
# Remove pacotes antigos que conflitam com o Docker oficial
sudo apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null

sudo apt update
sudo apt install -y ca-certificates curl gnupg

# Chave e repositório oficiais
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# ATENÇÃO no Linux Mint: use o codinome do Ubuntu correspondente
# (noble, jammy...), e não o do Mint (xia, virginia...).
CODINOME=$(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu $CODINOME stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io \
                    docker-buildx-plugin docker-compose-plugin
```

Para usar o Docker sem `sudo` (necessário para o kind funcionar):

```bash
sudo usermod -aG docker $USER
newgrp docker          # ou faça logout/login
docker run --rm hello-world
```

### 2. Task

```bash
sudo snap install task --classic
```

Sem snap, baixando o binário direto:

```bash
mkdir -p ~/.local/bin
curl -sL https://taskfile.dev/install.sh | sh -s -- -b ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"
```

### 3. kind e kubectl

Com o Task já instalado, uma tarefa cuida disso:

```bash
task tools
```

Manualmente, sem `sudo` (Linux x86_64):

```bash
mkdir -p ~/.local/bin

curl -Lo ~/.local/bin/kind https://kind.sigs.k8s.io/dl/v0.33.0/kind-linux-amd64
chmod +x ~/.local/bin/kind

curl -Lo ~/.local/bin/kubectl \
  "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ~/.local/bin/kubectl
```

Torne o diretório permanente no `PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

> O `Taskfile.yml` já injeta esse caminho no `PATH` por conta própria, então os comandos
> `task ...` funcionam mesmo sem este passo. Ele só é necessário para chamar `kind` e
> `kubectl` diretamente no terminal.

### 4. Go (opcional)

Só é preciso para compilar ou rodar o back-end fora do container — o build da imagem Docker
usa a sua própria cópia do Go.

```bash
curl -LO https://go.dev/dl/go1.26.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.5.linux-amd64.tar.gz
echo 'export PATH="$PATH:/usr/local/go/bin"' >> ~/.bashrc
source ~/.bashrc
go version
```

### Conferindo a instalação

```bash
task check
```

Saída esperada:

```
  ✔ docker   /usr/bin/docker
  ✔ kind     /home/voce/.local/bin/kind
  ✔ kubectl  /home/voce/.local/bin/kubectl
  ✔ daemon do Docker respondendo
```

Sem o Task, a conferência manual:

```bash
docker --version && docker compose version && docker info | head -5
kind --version
kubectl version --client
```

### Baixando as dependências do projeto

Nenhum passo manual é necessário: `task up` (ou `docker compose up --build`) baixa tudo
sozinho na primeira execução. As dependências Go são resolvidas dentro do build da imagem, e
o front-end não tem dependências.

Se quiser adiantar os downloads — por exemplo, para apresentar em uma rede ruim:

```bash
# Imagens que rodam em container (Compose e Kubernetes usam estas)
docker pull postgres:16-alpine
docker pull redis:7-alpine

# Imagens base dos builds
docker pull golang:1.26-alpine
docker pull alpine:3.20
docker pull nginx:1.27-alpine

# Dependências Go (Gin, goqu, pgx, go-redis) — só se for rodar o back-end
# fora do container; o build da imagem baixa por conta própria
cd backend && go mod download && cd ..
```

Depois disso, `task up` funciona sem rede. Para o Kubernetes o mesmo vale, já que as imagens
são carregadas no cluster a partir da máquina (veja a seção 7).

---

## 📋 1. Descrição do sistema

O sistema permite que um professor gerencie os alunos de uma disciplina e suas notas.

**Cadastro de alunos** — cadastrar, consultar, alterar e excluir. Cada aluno possui
`id`, `RA`, `nome`, `e-mail`, `curso` e `semestre`. Dois alunos não podem ter o mesmo RA.

**Cadastro de notas** — para cada aluno é possível lançar e alterar `P1` e `P2`.
A média e a situação são **sempre calculadas pelo back-end**, nunca aceitas do cliente:

```
Média = (P1 + P2) / 2

Média >= 6,0                  ->  Aprovado
Média >= 4,0 e Média < 6,0    ->  Exame
Média <  4,0                  ->  Reprovado
```

Alterar as notas recalcula média e situação e atualiza a listagem imediatamente.

**Validações** (aplicadas no front-end **e** repetidas no back-end):

| Campo | Regra |
|---|---|
| RA | obrigatório, somente dígitos, 5 a 8 posições, único no sistema |
| Nome | obrigatório, mínimo de 3 caracteres |
| E-mail | obrigatório, formato válido |
| Curso | obrigatório |
| Semestre | inteiro entre 1 e 10 |
| P1 / P2 | obrigatórias, numéricas, entre 0 e 10 |

A validação do front-end não é considerada suficiente: toda requisição é revalidada
pela API, que devolve os erros por campo em JSON e os exibe destacando cada input.

No back-end as regras são declaradas nas tags `binding` do Gin (incluindo uma validação
customizada `ra`, registrada em `internal/validate`), e os erros do validador são traduzidos
para mensagens em português indexadas pelo nome do campo — as mesmas chaves que o front-end
procura nos atributos `data-erro` do HTML.

---

## 🏗️ 2. Arquitetura da aplicação

```
                    ┌──────────────┐
                    │   Navegador  │
                    └──────┬───────┘
                           │ HTTP (:8080)
                    ┌──────▼────────────────┐
                    │  Front-end (nginx)    │
                    │  estáticos + proxy    │
                    └──────┬────────────────┘
                           │ /api  ->  backend-service:8080
                    ┌──────▼────────┐   2. GET   ┌─────────────┐
                    │   Back-end    │◄──────────►│    Redis    │
                    │   (Go / API)  │   4. SET   │    cache    │
                    └──────┬────────┘            └─────────────┘
                           │ 3. consulta em caso de cache MISS
                    ┌──────▼────────┐
                    │  PostgreSQL   │  (PersistentVolumeClaim)
                    └───────────────┘
```

Fluxo de uma consulta:

```
GET /api/alunos -> Redis -> cache MISS -> Banco -> grava no Redis (TTL 60s) -> resposta
GET /api/alunos -> Redis -> cache HIT  -> resposta
```

Toda escrita invalida o cache:

```
PUT /api/alunos/10 -> atualiza banco -> invalida Redis
                   -> próximo GET -> cache MISS -> consulta banco -> atualiza Redis
```

O front-end nunca fala com o banco ou com o Redis: consome **exclusivamente** a API REST.
O back-end alcança Redis e PostgreSQL pelos **Services do Kubernetes**
(`redis-service`, `database-service`) — nunca por `localhost`.

---

## 🛠️ 3. Tecnologias utilizadas

| Camada | Tecnologia |
|---|---|
| Front-end | HTML5, CSS3 e JavaScript (ES Modules), servido por **nginx 1.27-alpine** |
| Back-end | **Go 1.26** com o framework **Gin** (`github.com/gin-gonic/gin`) |
| Consultas SQL | **goqu** (`github.com/doug-martin/goqu/v9`), construtor de queries |
| Driver do banco | `github.com/jackc/pgx/v5` |
| Cliente de cache | `github.com/redis/go-redis/v9` |
| Banco de dados | **PostgreSQL 16** |
| Cache | **Redis 7** |
| Containers | Docker + Docker Compose |
| Orquestração | Kubernetes (kind) |

### Estrutura do projeto

```
sistema-academico/
├── docker-compose.yml           # ambiente local completo
├── kind-config.yaml             # cluster kind, publica o NodePort 30080 em localhost:8080
├── frontend/
│   ├── Dockerfile
│   ├── nginx/default.conf.template
│   ├── index.html
│   ├── css/style.css
│   └── js/{app,api,estado,alunos,notas,ui}.js
├── backend/
│   ├── Dockerfile
│   ├── cmd/api/main.go
│   └── internal/{config,store,cache,httpapi,model,validate}/
├── database/
│   └── init.sql                 # referência do esquema
└── kubernetes/
    ├── configmap.yaml           secret.yaml
    ├── frontend-deployment.yaml frontend-service.yaml
    ├── backend-deployment.yaml  backend-service.yaml
    ├── redis-deployment.yaml    redis-service.yaml
    └── database-deployment.yaml database-service.yaml  database-pvc.yaml
```

---

## ▶️ 4. Como executar a aplicação localmente

Requisitos: Docker e Docker Compose.

```bash
docker compose up --build -d
docker compose ps
```

> Equivalente com o Taskfile: `task up`



Sobem quatro containers: `sa-frontend`, `sa-backend`, `sa-redis` e `sa-database`.

| Endereço | O quê |
|---|---|
| http://localhost:8080 | aplicação |
| http://localhost:8081 | API direta, para testes com `curl` |

Encerrar:

```bash
docker compose down          # mantém os dados
docker compose down -v       # remove também o volume do Postgres
```

> Equivalentes: `task down` e `task clean`

### Executar o back-end fora do container (opcional)

```bash
docker compose up -d database redis
cd backend
DB_HOST=localhost DB_PASSWORD=postgres REDIS_HOST=localhost go run ./cmd/api
```

---

## 🐳 5. Como construir as imagens Docker

```bash
docker build -t sistema-academico/backend:1.0  ./backend
docker build -t sistema-academico/frontend:1.0 ./frontend

docker images | grep sistema-academico
```

> Equivalente com o Taskfile: `task build`

O back-end usa build multi-estágio: compila com `golang:1.26-alpine` e entrega um
binário estático em `alpine:3.20`, executado por usuário não-root.
Redis e PostgreSQL usam as imagens oficiais e por isso não possuem Dockerfile.

---

## ☸️ 6. Como iniciar o ambiente Kubernetes

Ambiente utilizado: **kind** (Kubernetes rodando dentro do Docker).

Instalação das ferramentas (Linux x86_64), sem precisar de `sudo`:

```bash
mkdir -p ~/.local/bin

curl -Lo ~/.local/bin/kind https://kind.sigs.k8s.io/dl/v0.33.0/kind-linux-amd64
chmod +x ~/.local/bin/kind

curl -Lo ~/.local/bin/kubectl \
  "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ~/.local/bin/kubectl

export PATH="$HOME/.local/bin:$PATH"   # acrescente ao ~/.bashrc para tornar permanente

kind --version && kubectl version --client
```

Criação do cluster:

```bash
kind create cluster --config kind-config.yaml
kubectl cluster-info
kubectl get nodes
```

> Equivalente com o Taskfile: `task k8s:up` (cria o cluster se não existir, ou
> apenas o religa se estiver parado)

> O `kind-config.yaml` publica o NodePort 30080 na porta 8080 da máquina. Se o
> `docker compose` estiver no ar, encerre-o antes (`docker compose down`) para
> liberar essa porta.

---

## 🚀 7. Como realizar o deploy

> Equivalente com o Taskfile: `task k8s:up` (faz o build, carrega as imagens, aplica os
> manifests e espera o rollout — tudo em um comando)

O nó do kind tem seu próprio armazenamento de imagens, separado do Docker da máquina. As
quatro imagens precisam ser carregadas para dentro dele antes do `apply` — inclusive as
oficiais, senão o kubelet tenta baixá-las do Docker Hub a cada cluster novo:

```bash
for img in sistema-academico/backend:1.0 sistema-academico/frontend:1.0 \
           postgres:16-alpine redis:7-alpine; do
  docker save --platform linux/amd64 "$img" -o /tmp/img.tar
  kind load image-archive /tmp/img.tar --name sistema-academico
done
rm -f /tmp/img.tar
```

> **Por que não `kind load docker-image`?** O Docker recente armazena as imagens no image
> store do containerd, que guarda a lista multi-plataforma completa. O `kind load
> docker-image` importa com `--all-platforms` e falha com `content digest ... not found`,
> porque só a variante `linux/amd64` foi de fato baixada. Exportar uma plataforma
> explicitamente com `docker save --platform` resolve. Em versões mais antigas do Docker
> (sem a flag `--platform`), `kind load docker-image <imagem> --name sistema-academico`
> funciona normalmente.

Todos os Deployments usam `imagePullPolicy: IfNotPresent`, então usam a imagem já presente
no nó em vez de procurar um registro remoto.

Aplicando os manifests:

```bash
kubectl apply -f kubernetes/

kubectl rollout status deployment/database
kubectl rollout status deployment/redis
kubectl rollout status deployment/backend
kubectl rollout status deployment/frontend
```

### Reimplantar depois de alterar o código

As imagens mantêm a mesma tag (`1.0`), então o Kubernetes não percebe sozinho que elas
mudaram. É preciso reconstruir, recarregar e reiniciar:

```bash
docker build -t sistema-academico/backend:1.0 ./backend
docker save --platform linux/amd64 sistema-academico/backend:1.0 -o /tmp/img.tar
kind load image-archive /tmp/img.tar --name sistema-academico
kubectl rollout restart deployment/backend
```

> Equivalente com o Taskfile: `task k8s:up` (é idempotente — pode rodar quantas vezes quiser)

Remover tudo:

```bash
kubectl delete -f kubernetes/                      # remove os recursos, mantém o cluster
kind delete cluster --name sistema-academico       # remove o cluster inteiro
```

> Equivalentes: `task k8s:down` e `task k8s:delete`

---

## 🌐 8. Como acessar o sistema

O `frontend-service` é um **NodePort 30080**, publicado pelo `kind-config.yaml` na
porta 8080 da máquina:

```
http://localhost:8080
```

Sem o `extraPortMappings` (ou em outro ambiente), use:

```bash
kubectl port-forward service/frontend-service 8080:80
```

Para testar a API diretamente, sem passar pelo nginx:

```bash
kubectl port-forward service/backend-service 8081:8080
curl -i localhost:8081/api/alunos
```

---

## 📦 9. Como verificar os Pods

```bash
kubectl get pods
kubectl get pods -o wide
kubectl describe pod -l app=backend
kubectl logs -l app=backend --tail=50 -f
```

> Equivalentes: `task k8s:status` e `task k8s:logs`

Esperado: 1 Pod de `frontend`, 1 de `redis`, 1 de `database` e **2 de `backend`**.

---

## 🔌 10. Como verificar os Services

```bash
kubectl get services
kubectl describe service backend-service

# Os endpoints do backend-service devem listar os IPs das duas réplicas:
kubectl get endpoints backend-service
```

| Service | Tipo | Porta |
|---|---|---|
| `frontend-service` | NodePort | 80 -> 30080 |
| `backend-service` | ClusterIP | 8080 |
| `redis-service` | ClusterIP | 6379 |
| `database-service` | ClusterIP | 5432 |

> Equivalente com o Taskfile: `task test:k8s`

---

## 📊 11. Como verificar os Deployments

```bash
kubectl get deployments
kubectl describe deployment backend
kubectl get configmap sistema-academico-config -o yaml
kubectl get secret sistema-academico-secret -o yaml
kubectl get pvc
```
> Equivalente com o Taskfile: `task test:k8s`

---

## 🗃️ 12. Como verificar o Redis

```bash
REDIS=$(kubectl get pod -l app=redis -o name)

kubectl exec -it $REDIS -- redis-cli ping          # PONG
kubectl exec -it $REDIS -- redis-cli KEYS '*'      # alunos:all, aluno:1, ...
kubectl exec -it $REDIS -- redis-cli TTL alunos:all
kubectl exec -it $REDIS -- redis-cli MONITOR       # acompanha GET/SET/DEL ao vivo
```
> Equivalente com o Taskfile: `task test:redis` (faz MISS, HIT, TTL e invalidação em sequência)

Demonstração de **MISS -> HIT -> invalidação** (com `kubectl port-forward service/backend-service 8081:8080`):

```bash
curl -sI localhost:8081/api/alunos | grep X-Cache     # X-Cache: MISS
curl -sI localhost:8081/api/alunos | grep X-Cache     # X-Cache: HIT

kubectl exec -it $REDIS -- redis-cli TTL alunos:all   # 60 -> decrescendo

curl -X PUT localhost:8081/api/alunos/1 -H 'Content-Type: application/json' \
     -d '{"ra":"12345","nome":"João Silva","email":"joao@email.com","curso":"ADS","semestre":5}'

curl -sI localhost:8081/api/alunos | grep X-Cache     # X-Cache: MISS de novo
```

O header `X-Cache` é emitido pela própria API a cada consulta cacheável, e cada
requisição também é registrada nos logs do Pod (`kubectl logs -l app=backend`).

---

## 📈 13. Como escalar o back-end

O Deployment já sobe com **2 réplicas** (seção 19 do enunciado):

```bash
kubectl scale deployment backend --replicas=3
kubectl get pods -l app=backend        # três Pods em Running
kubectl get deployment backend         # READY 3/3
```

A aplicação continua funcionando normalmente durante e depois do escalonamento: o
`backend-service` distribui as requisições entre os Pods, e o estado fica no banco e
no Redis — os Pods do back-end não guardam estado.

Voltar ao original:

```bash
kubectl scale deployment backend --replicas=2
```
> Equivalente com o Taskfile: `task test:scale` (escala para 3, confere e volta para 2)

---

## 💾 Persistência do banco

O PostgreSQL grava em um PersistentVolumeClaim de 1Gi. Para demonstrar que os dados
sobrevivem à recriação do Pod:

```bash
# 1. cria um registro
curl -X POST localhost:8081/api/alunos -H 'Content-Type: application/json' \
     -d '{"ra":"777777","nome":"Aluno Persistente","email":"p@email.com","curso":"ADS","semestre":1}'

# 2. destrói o Pod do banco — o Deployment cria outro no lugar
kubectl delete pod -l app=database
kubectl rollout status deployment/database

# 3. o registro continua lá
curl -s localhost:8081/api/alunos | grep 777777
```
> Equivalente com o Taskfile: `task test:persistencia`

---

## 🎬 Roteiro de demonstração (seção 22 do enunciado)

Cada item exigido na entrega tem um comando pronto. Todos imprimem o resultado esperado ao
lado do obtido, para conferência durante a apresentação.

| Item da entrega | Comando | O que demonstra |
|---|---|---|
| Front-end (CRUD, notas, média, situação) | `task test:responsivo` e uso da tela | cadastro, consulta, alteração, exclusão, notas e situação |
| Validação de dados | `task test:validacao` | erros por campo devolvidos pela API |
| API (chamadas REST) | `task test:api` | 200, 201, 204, 400, 404 e 409 em cada rota |
| Docker | `task test:docker` | containers em execução e imagens utilizadas |
| Kubernetes | `task test:k8s` | Pods, Deployments, Services, ConfigMap, Secret e PVC |
| Redis | `task test:redis` | cache MISS, cache HIT, TTL e invalidação |
| Escalabilidade | `task test:scale` | `kubectl scale --replicas=3` com a aplicação no ar |
| Persistência | `task test:persistencia` | dado sobrevive à recriação do Pod do banco |
| Responsividade | `task test:responsivo` | as quatro resoluções obrigatórias |

Bateria completa, com o ambiente já no Kubernetes:

```bash
task k8s:up
task test:all
```

As seções seguintes trazem os mesmos comandos escritos um a um, caso prefira executá-los
manualmente durante a apresentação.

---

## 🔗 API REST

Base: `/api`. Todas as respostas são JSON.

### Alunos

| Método | Rota | Descrição | Status |
|---|---|---|---|
| GET | `/api/alunos` | lista todos os alunos com suas notas (**com cache**) | 200 |
| GET | `/api/alunos/{id}` | consulta um aluno (**com cache**) | 200, 404 |
| POST | `/api/alunos` | cadastra | 201, 400, 409 |
| PUT | `/api/alunos/{id}` | altera | 200, 400, 404, 409 |
| DELETE | `/api/alunos/{id}` | exclui (remove as notas junto) | 204, 404 |

### Notas

| Método | Rota | Descrição | Status |
|---|---|---|---|
| GET | `/api/alunos/{id}/notas` | notas do aluno | 200, 404 |
| POST | `/api/alunos/{id}/notas` | lança as notas | 201, 400, 404, 409 |
| PUT | `/api/notas/{id}` | altera as notas | 200, 400, 404 |
| DELETE | `/api/notas/{id}` | exclui o lançamento | 204, 404 |

`GET /health` completa a lista, alimentando as probes do Kubernetes.

Códigos utilizados: **200** OK, **201** Created, **204** No Content, **400** Bad Request,
**404** Not Found, **409** Conflict (RA duplicado ou notas já lançadas) e
**500** Internal Server Error.

### Exemplos

```bash
# Listar
curl -i localhost:8081/api/alunos

# Cadastrar
curl -i -X POST localhost:8081/api/alunos -H 'Content-Type: application/json' \
  -d '{"ra":"123456","nome":"João da Silva","email":"joao@email.com",
       "curso":"Análise e Desenvolvimento de Sistemas","semestre":4}'

# Lançar notas
curl -i -X POST localhost:8081/api/alunos/1/notas -H 'Content-Type: application/json' \
  -d '{"p1":8.0,"p2":7.0}'

# Alterar notas
curl -i -X PUT localhost:8081/api/notas/1 -H 'Content-Type: application/json' \
  -d '{"p1":5.0,"p2":4.0}'

# Excluir aluno
curl -i -X DELETE localhost:8081/api/alunos/1
```

Resposta de sucesso:

```json
{
  "id": 1, "ra": "12345", "nome": "João Silva", "email": "joao.silva@email.com",
  "curso": "Análise e Desenvolvimento de Sistemas", "semestre": 4,
  "nota": { "id": 1, "aluno_id": 1, "p1": 8, "p2": 7, "media": 7.5, "situacao": "Aprovado" }
}
```

Resposta de erro de validação (HTTP 400):

```json
{
  "erro": "Existem campos inválidos no formulário.",
  "campos": {
    "ra": "RA inválido: informe de 5 a 8 dígitos.",
    "email": "Formato de e-mail inválido (ex.: joao@email.com).",
    "semestre": "O semestre deve ser um número inteiro entre 1 e 10."
  }
}
```

---

## 📱 Responsividade

A interface foi testada nas quatro resoluções obrigatórias e em uma quinta intermediária:

| Dispositivo | Resolução | Comportamento |
|---|---|---|
| Smartphone | 375 × 667 | menu hambúrguer; tabela vira cartões; formulário em coluna única |
| Smartphone | 414 × 896 | idem |
| Tablet | 768 × 1024 | menu hambúrguer; cartões; formulário em duas colunas |
| Desktop | 1920 × 1080 | menu horizontal; tabela completa com 10 colunas |
| Intermediária | 1024 × 768 | menu horizontal; tabela rolando dentro do próprio contêiner |

Em nenhuma das larguras a página apresenta rolagem horizontal, texto cortado ou
elementos sobrepostos. Abaixo de 768px a tabela é convertida em cartões, com cada
valor rotulado (`data-label` + `::before`), e os botões de ação ocupam a largura
total para facilitar o toque. Acima disso, a tabela rola dentro do seu próprio
contêiner, sem arrastar a página junto. Tipografia e espaçamentos usam `clamp()`
e grids `auto-fit`/`minmax`, então o layout também se comporta em resoluções não
previstas.

---

## ⚙️ Variáveis de ambiente

| Variável | Origem no Kubernetes | Padrão | Descrição |
|---|---|---|---|
| `PORT` | — | `8080` | porta HTTP do back-end |
| `DB_HOST` | ConfigMap | `localhost` | host do PostgreSQL |
| `DB_PORT` | ConfigMap | `5432` | porta do PostgreSQL |
| `DB_NAME` | ConfigMap | `sistema_academico` | nome do banco |
| `DB_USER` | ConfigMap | `postgres` | usuário do banco |
| `DB_PASSWORD` | **Secret** | — | senha do banco |
| `REDIS_HOST` | ConfigMap | `localhost` | host do Redis |
| `REDIS_PORT` | ConfigMap | `6379` | porta do Redis |
| `CACHE_TTL_SECONDS` | ConfigMap | `60` | TTL das chaves de cache |
| `SEED_ON_EMPTY` | — | `true` | cria alunos de exemplo se o banco estiver vazio |
| `GIN_MODE` | ConfigMap | `release` (na imagem) | `release` silencia os avisos de debug do Gin |
| `BACKEND_URL` | ConfigMap | `http://backend:8080` | destino do proxy `/api` no nginx |
